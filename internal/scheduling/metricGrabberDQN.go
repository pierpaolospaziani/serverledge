package scheduling

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"strconv"

	"github.com/grussorusso/serverledge/internal/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const INFLUXDB = "[INFLUXDB]:"

type DQNStats struct {
	Exec	[]float64	// list of seconds in which exec append
	Cloud	[]float64	// list of seconds in which offload cloud append
	Edge	[]float64	// list of seconds in which offload edge append
	Drop	[]float64	// list of seconds in which drop append

	Reward 	[]float64	// list of rewards obtained
	Cost 	[]float64	// list of costs

	// counts how many times an action has been taken for that class
	// LOCAL(0)-CLOUD(1)-EDGE(2)-DROP(3)
	Standard 	[]int
	Critical1 	[]int
	Critical2 	[]int
	Batch 		[]int

	ResponseTime 		[]float64
	IsWarmStart  		[]bool
	InitTime 	 		[]float64
	Duration 	 		[]float64
	OffloadLatencyCloud []float64
	OffloadLatencyEdge 	[]float64
}

var stats DQNStats

var initTime time.Time

var updateEvery = config.GetInt(config.DQN_STORE_STATS_EVERY, 3600)
var updateRound = 0

// metricGrabberDQN encapsulates the InfluxDB client and configuration
type metricGrabberDQN struct {
	client   influxdb2.Client
	org      string
	bucket   string
	writeAPI api.WriteAPIBlocking
}


func EmptyStats() DQNStats{
	stats = DQNStats{
	    Exec:      			 []float64{},
	    Cloud:     			 []float64{},
	    Edge:      			 []float64{},
	    Drop:      			 []float64{},
	    Reward:    			 []float64{},
	    Cost:      			 []float64{},
	    Standard:  			 []int{0,0,0,0},
	    Critical1: 			 []int{0,0,0,0},
	    Critical2: 			 []int{0,0,0,0},
	    Batch:     			 []int{0,0,0,0},
	    ResponseTime: 		 []float64{},
		IsWarmStart:		 []bool{},
		InitTime: 			 []float64{},
		Duration: 			 []float64{},
		OffloadLatencyCloud: []float64{},
		OffloadLatencyEdge:  []float64{},
	}
	return stats
}


// Initializes and returns a new metricGrabberDQN instance
func InitMG() *metricGrabberDQN {
	stats = EmptyStats()

	initTime = time.Now()

	org := config.GetString(config.STORAGE_DB_ORGNAME, "serverledge")
	url := config.GetString(config.STORAGE_DB_ADDRESS, "http://localhost:8086")
	token := config.GetString(config.STORAGE_DB_TOKEN, "serverledge")
	bucket := config.GetString(config.STORAGE_DB_BUCKET, "dqn")

	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)
	return &metricGrabberDQN{
		client:   client,
		org:      org,
		bucket:   bucket,
		writeAPI: writeAPI,
	}
}


func (mg *metricGrabberDQN) addStats(r *scheduledRequest, dropped bool) {
	elapsedTime := r.Arrival.Sub(initTime)
	elapsedTimeInSeconds := float64(elapsedTime.Seconds())
	if dropped {
		stats.Drop = append(stats.Drop, elapsedTimeInSeconds)
	} else {
		switch r.ExecReport.SchedAction {
		case "O_C":
		    stats.Cloud = append(stats.Cloud, float64(elapsedTime.Seconds()))
		case "O_E":
		    stats.Edge = append(stats.Edge, float64(elapsedTime.Seconds()))
		default:
		    stats.Exec = append(stats.Exec, float64(elapsedTime.Seconds()))
		}
	}

	if !dropped && (r.ExecReport.ResponseTime <= r.ClassService.MaximumResponseTime || r.ClassService.MaximumResponseTime == -1) {
		stats.Reward = append(stats.Reward, r.ClassService.Utility)
	} else {
		stats.Reward = append(stats.Reward, 0)
	}

	stats.Cost = append(stats.Cost, r.ExecReport.Cost)

	switch r.ClassService.Name {
    case "batch":
        updateClassStats(&stats.Batch, dropped, r.ExecReport.SchedAction)
    case "critical-1":
        updateClassStats(&stats.Critical1, dropped, r.ExecReport.SchedAction)
    case "critical-2":
        updateClassStats(&stats.Critical2, dropped, r.ExecReport.SchedAction)
    default:
        updateClassStats(&stats.Standard, dropped, r.ExecReport.SchedAction)
    }

    if !dropped {
    	stats.ResponseTime = append(stats.ResponseTime, r.ExecReport.ResponseTime)
    	stats.IsWarmStart = append(stats.IsWarmStart, r.ExecReport.IsWarmStart)
    	stats.InitTime = append(stats.InitTime, r.ExecReport.InitTime)
    	stats.Duration = append(stats.Duration, r.ExecReport.Duration)
		if r.ExecReport.SchedAction == "O_C"{
    		stats.OffloadLatencyCloud = append(stats.OffloadLatencyCloud, r.ExecReport.OffloadLatencyCloud)
		} else if r.ExecReport.SchedAction == "O_E"{
	    	stats.OffloadLatencyEdge = append(stats.OffloadLatencyEdge, r.ExecReport.OffloadLatencyEdge)
		}
    }

    // add stats to InfluxDB every 'updateEvery'
    if elapsedTimeInSeconds - float64(updateEvery*updateRound) > float64(updateEvery) {
    	mg.WriteJSON()
		stats = EmptyStats()
    	updateRound++
    }
}


func updateClassStats(slice *[]int, dropped bool, schedAction string) {
    index := 3
    if !dropped {
        switch schedAction {
        case "O_C":
            index = 1
        case "O_E":
            index = 2
        default:
            index = 0
        }
    }
    (*slice)[index]++
}


// Writes stats as a JSON object to InfluxDB
func (mg *metricGrabberDQN) WriteJSON() {
	parts := []interface{}{
        stats.Exec, stats.Cloud, stats.Edge, stats.Drop,
        stats.Reward, stats.Cost, stats.Standard, stats.Critical1,
        stats.Critical2, stats.Batch, stats.ResponseTime, stats.IsWarmStart,
        stats.InitTime, stats.Duration, stats.OffloadLatencyCloud, stats.OffloadLatencyEdge,
    }

    for i, part := range parts {
		// Convert DQNStats to JSON string
		jsonData, err := json.Marshal(part)
        if err != nil {
            log.Fatalf("%s Error marshalling JSON part %d: %v\n", INFLUXDB, i, err)
        }

		// Create a new data point
		point := influxdb2.NewPointWithMeasurement("dqn_stats").
            AddTag("part", "part_" + strconv.Itoa(i)).
            AddField("json_data", string(jsonData)).
            SetTime(time.Now().UTC())

		// Write the point to InfluxDB
		err = mg.writeAPI.WritePoint(context.Background(), point)
        if err != nil {
            log.Fatalf("%s Error writing point to InfluxDB for part %d: %v\n", INFLUXDB, i, err)
        }
    }
	log.Println(INFLUXDB, "Statistics successfully written to InfluxDB")
}


// Closes the InfluxDB client connection
func (mg *metricGrabberDQN) Close() {
	mg.client.Close()
}