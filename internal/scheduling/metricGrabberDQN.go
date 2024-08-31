package scheduling

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"strconv"
    "unicode"
    "os"

	"github.com/grussorusso/serverledge/internal/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const INFLUXDB = "[INFLUXDB]:"

type ClassPenalties struct {
    Name			string  `json:"name"`
    DeadlinePenalty	float64 `json:"deadlinePenalty"`
    DropPenalty 	float64 `json:"dropPenalty"`
}

type DQNStats struct {
	Exec	[]float64	// list of seconds in which exec append
	Cloud	[]float64	// list of seconds in which offload cloud append
	Edge	[]float64	// list of seconds in which offload edge append
	Drop	[]float64	// list of seconds in which drop append

	Reward 			[]float64	// list of rewards obtained
	DeadlinePenalty	[]float64	// list of deadline penalties
	DropPenalty 	[]float64	// list of drop penalties
	Cost 			[]float64	// list of costs

	// counts how many times an action has been taken for that class
	// LOCAL(0)-CLOUD(1)-EDGE(2)-DROP(3)
	Standard 	[]int
	Critical1 	[]int
	Critical2 	[]int
	Batch 		[]int

	// stats for function (f1,f2,f3,f4,f5)
	ResponseTime 		[][]float64
	IsWarmStart  		[][]int 	// [[f1_warm,f1_cold], ... ]
	InitTime 	 		[][]float64
	Duration 	 		[][]float64
	OffloadLatencyCloud [][]float64
	OffloadLatencyEdge 	[][]float64
}

var penalties []ClassPenalties

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
	    DeadlinePenalty:	 []float64{},
	    DropPenalty:		 []float64{},
	    Cost:      			 []float64{},
	    Standard:  			 []int{0,0,0,0},
	    Critical1: 			 []int{0,0,0,0},
	    Critical2: 			 []int{0,0,0,0},
	    Batch:     			 []int{0,0,0,0},
	    ResponseTime: 		 [][]float64{},
		IsWarmStart:		 [][]int{{0,0},{0,0},{0,0},{0,0},{0,0}},
		InitTime: 			 [][]float64{},
		Duration: 			 [][]float64{},
		OffloadLatencyCloud: [][]float64{},
		OffloadLatencyEdge:  [][]float64{},
	}
	return stats
}


// Initializes and returns a new metricGrabberDQN instance
func InitMG() *metricGrabberDQN {
	stats = EmptyStats()

	initTime = time.Now()

	// retrieve penalties from json classes file
	file, err := os.Open("serverledge-classes.json")
    if err != nil {
	    log.Println("%s Error opening classes file\n", INFLUXDB)
	    panic(err)
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&penalties)
    if err != nil {
	    log.Println("%s Error during JSON decode\n", INFLUXDB)
	    panic(err)
    }

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
		stats.DropPenalty = append(stats.Reward, 0)
		stats.DeadlinePenalty = append(stats.Reward, 0)
	} else {
		stats.Reward = append(stats.Reward, 0)
		for _, penalty := range penalties {
			log.Println(penalty.Name, r.ClassService.Name)
			if penalty.Name == r.ClassService.Name {
				log.Println(true)
				if dropped {
					stats.DropPenalty = append(stats.Reward, penalty.DropPenalty)
					stats.DeadlinePenalty = append(stats.Reward, 0)
				} else {
					stats.DropPenalty = append(stats.Reward, 0)
					stats.DeadlinePenalty = append(stats.Reward, penalty.DeadlinePenalty)
				}
				break
			}
		}
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
		var numStr string
		for _, char := range r.Fun.Name {
	        if unicode.IsDigit(char) {
	            numStr += string(char)
	        }
	    }
	    index, err := strconv.Atoi(numStr)
	    if err != nil {
	    	log.Fatalf("%s Error during function number conversion:%v\n", INFLUXDB, err)
	    }
	    index--

    	stats.ResponseTime[index] = append(stats.ResponseTime[index], r.ExecReport.ResponseTime)
    	if r.ExecReport.IsWarmStart {
    		stats.IsWarmStart[index][0]++
    	} else {
    		stats.IsWarmStart[index][1]++
    	}
    	stats.InitTime[index] = append(stats.InitTime[index], r.ExecReport.InitTime)
    	stats.Duration[index] = append(stats.Duration[index], r.ExecReport.Duration)
		if r.ExecReport.SchedAction == "O_C"{
			stats.OffloadLatencyCloud[index] = append(stats.OffloadLatencyCloud[index], r.ExecReport.OffloadLatencyCloud)
		} else if r.ExecReport.SchedAction == "O_E"{
			stats.OffloadLatencyEdge[index] = append(stats.OffloadLatencyEdge[index], r.ExecReport.OffloadLatencyEdge)
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