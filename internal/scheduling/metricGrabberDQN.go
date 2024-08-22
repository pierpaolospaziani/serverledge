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
	exec		[]int
	cloud		[]int
	edge		[]int
	drop		[]int

	reward 		[]int
	reward_cost []int
	cost 		[]int

	standard 	[]int
	critical_1 	[]int
	critical_2 	[]int
	batch 		[]int
}

// metricGrabberDQN encapsulates the InfluxDB client and configuration
type metricGrabberDQN struct {
	client   influxdb2.Client
	org      string
	bucket   string
	writeAPI api.WriteAPIBlocking
}

// Initializes and returns a new metricGrabberDQN instance
func InitMG() *metricGrabberDQN {
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

// WriteJSON writes a JSON object to InfluxDB
func (mg *metricGrabberDQN) WriteJSON(stats DQNStats) {
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