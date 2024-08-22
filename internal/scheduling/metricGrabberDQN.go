package scheduling

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/grussorusso/serverledge/internal/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const INFLUXDB = "[INFLUXDB]:"

type DQNStats struct {
	exec		[]float64
	cloud		[]float64
	edge		[]float64
	drop		[]float64

	reward 		[]float64
	cost 		[]float64

	standard 	[]int
	critical1 	[]int
	critical2 	[]int
	batch 		[]int
}

var stats DQNStats

var initTime time.Time

// metricGrabberDQN encapsulates the InfluxDB client and configuration
type metricGrabberDQN struct {
	client   influxdb2.Client
	org      string
	bucket   string
	writeAPI api.WriteAPIBlocking
}


// Initializes and returns a new metricGrabberDQN instance
func InitMG() *metricGrabberDQN {
	stats = DQNStats{
	    exec:      []float64{},
	    cloud:     []float64{},
	    edge:      []float64{},
	    drop:      []float64{},
	    reward:    []float64{},
	    cost:      []float64{},
	    standard:  []int{0,0,0,0},
	    critical1: []int{0,0,0,0},
	    critical2: []int{0,0,0,0},
	    batch:     []int{0,0,0,0},
	}

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
	time := r.Arrival.Sub(initTime)
	log.Println("DECISION:",r.ExecReport.SchedAction, "DROPPED:", dropped)
	if dropped {
		stats.drop = append(stats.drop, float64(time.Seconds()))
	} else {
		switch r.ExecReport.SchedAction {
		case "O_C":
		    stats.cloud = append(stats.cloud, float64(time.Seconds()))
		case "O_E":
		    stats.edge = append(stats.edge, float64(time.Seconds()))
		default:
		    stats.exec = append(stats.exec, float64(time.Seconds()))
		}
	}

	if !dropped && (r.ExecReport.ResponseTime <= r.ClassService.MaximumResponseTime || r.ClassService.MaximumResponseTime == -1) {
		stats.reward = append(stats.reward, r.ClassService.Utility)
	} else{
		stats.reward = append(stats.reward, 0)
	}

	stats.cost = append(stats.cost, r.ExecReport.Cost)

	switch r.ClassService.Name {
    case "batch":
        updateClassStats(stats.batch, dropped, r.ExecReport.SchedAction)
    case "critical1":
        updateClassStats(stats.critical1, dropped, r.ExecReport.SchedAction)
    case "critical2":
        updateClassStats(stats.critical2, dropped, r.ExecReport.SchedAction)
    default:
        updateClassStats(stats.standard, dropped, r.ExecReport.SchedAction)
    }
}


func updateClassStats(slice []int, dropped bool, schedAction string) {
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
    slice[index]++
}


// WriteJSON writes a JSON object to InfluxDB
func (mg *metricGrabberDQN) WriteJSON() {
	// Convert DQNStats to JSON string
	jsonData, err := json.Marshal(stats)
	if err != nil {
		log.Fatalf("%s Error marshalling JSON: %v\n", INFLUXDB, err)
	}

	// Create a new data point
	point := influxdb2.NewPointWithMeasurement("dqn_stats").
		AddTag("new_data", "new_data").
		AddField("json_data", string(jsonData)).
		SetTime(time.Now().UTC())

	// Write the point to InfluxDB
	err = mg.writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		log.Fatalf("%s Error writing point to InfluxDB: %v\n", INFLUXDB, err)
	}

	log.Println(INFLUXDB, "Statistics successfully written to InfluxDB")
}


// Close closes the InfluxDB client connection
func (mg *metricGrabberDQN) Close() {
	mg.client.Close()
}