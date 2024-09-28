package scheduling

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"strconv"
    "unicode"
    "os"
	"sync"

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
	Exec  []float64	// list of seconds in which exec append
	Cloud []float64	// list of seconds in which offload cloud append
	Edge  []float64	// list of seconds in which offload edge append
	Drop  []float64	// list of seconds in which drop append

	Reward 			[]float64	// list of rewards obtained
	DeadlinePenalty	[]float64	// list of deadline penalties
	DropPenalty 	[]float64	// list of drop penalties
	Cost 			[]float64	// list of costs

	// counts how many times an action has been taken for that class
	// LOCAL(0)-CLOUD(1)-EDGE(2)-DROP(3)
	Standard  []int
	Critical1 []int
	Critical2 []int
	Batch 	  []int

	// counts how many times an action has been taken for that function
	// LOCAL(0)-CLOUD(1)-EDGE(2)-DROP(3)
	f1 []int
	f2 []int
	f3 []int
	f4 []int
	f5 []int

	// stats for function (f1,f2,f3,f4,f5)
	ResponseTime 		[][]float64
	IsWarmStart  		[][]int 	// [[f1_warm,f1_cold], ... ]
	InitTime 	 		[][]float64
	Duration 	 		[][]float64
	OffloadLatencyCloud [][]float64
	OffloadLatencyEdge 	[][]float64

	// [utility,deadlinePenalty,dropPenalty] per action
	UPExec	[]float64
	UPCloud	[]float64
	UPEdge	[]float64

	// [utility,deadlinePenalty,dropPenalty] per class
	UPStandard 	[]float64
	UPCritical1 []float64
	UPCritical2 []float64
	UPBatch 	[]float64

	// [utility,deadlinePenalty,dropPenalty] per function
	UPf1 []float64
	UPf2 []float64
	UPf3 []float64
	UPf4 []float64
	UPf5 []float64
}

var penaltyMap map[string][]float64

var stats DQNStats
var muStats sync.Mutex

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
	    f1:  			 	 []int{0,0,0,0},
	    f2: 			 	 []int{0,0,0,0},
	    f3: 			 	 []int{0,0,0,0},
	    f4:     			 []int{0,0,0,0},
	    f5:     			 []int{0,0,0,0},
	    ResponseTime: 		 [][]float64{{},{},{},{},{}},
		IsWarmStart:		 [][]int{{0,0},{0,0},{0,0},{0,0},{0,0}},
		InitTime: 			 [][]float64{{},{},{},{},{}},
		Duration: 			 [][]float64{{},{},{},{},{}},
		OffloadLatencyCloud: [][]float64{{},{},{},{},{}},
		OffloadLatencyEdge:  [][]float64{{},{},{},{},{}},
		UPExec:      		 []float64{0,0,0},
	    UPCloud:     		 []float64{0,0,0},
	    UPEdge:      		 []float64{0,0,0},
	    UPStandard:      	 []float64{0,0,0},
	    UPCritical1:    	 []float64{0,0,0},
	    UPCritical2: 		 []float64{0,0,0},
	    UPBatch:	 		 []float64{0,0,0},
	    UPf1:	      		 []float64{0,0,0},
	    UPf2:	      	 	 []float64{0,0,0},
	    UPf3:	    	 	 []float64{0,0,0},
	    UPf4:	 		 	 []float64{0,0,0},
	    UPf5:		 		 []float64{0,0,0},
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
    var penalties []ClassPenalties
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&penalties)
    if err != nil {
	    log.Println("%s Error during JSON decode\n", INFLUXDB)
	    panic(err)
    }
    penaltyMap = make(map[string][]float64)
    for _, penalty := range penalties {
        penaltyMap[penalty.Name] = []float64{penalty.DeadlinePenalty, penalty.DropPenalty}
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


func (mg *metricGrabberDQN) addStats(r *scheduledRequest, actionDrop bool, offloadDrop bool) {
	muStats.Lock()
	defer muStats.Unlock()

	dropped := false
	if actionDrop || offloadDrop {
		dropped = true
	}
	outOfTime := false
	if !dropped && r.ExecReport.ResponseTime > r.ClassService.MaximumResponseTime {
		outOfTime = true
	}

	elapsedTime := r.Arrival.Sub(initTime)
	elapsedTimeInSeconds := float64(elapsedTime.Seconds())
	
	// actionDrop (and not dropped) cause i want the chosen action not the outcome of the request
	if actionDrop {
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

	// 'dropped' cause i want the outcome of the request
	if !dropped && !outOfTime {
		stats.Reward = append(stats.Reward, r.ClassService.Utility)
		stats.DropPenalty = append(stats.DropPenalty, 0)
		stats.DeadlinePenalty = append(stats.DeadlinePenalty, 0)
	} else {
		stats.Reward = append(stats.Reward, 0)
		penalties, exists := penaltyMap[r.ClassService.Name]
		if !exists {
			penalties = []float64{0,0} // for default class
		}
		if dropped {
			stats.DropPenalty = append(stats.DropPenalty, penalties[0])
			stats.DeadlinePenalty = append(stats.DeadlinePenalty, 0)
		} else {
			stats.DropPenalty = append(stats.DropPenalty, 0)
			stats.DeadlinePenalty = append(stats.DeadlinePenalty, penalties[1])
		}
	}

	stats.Cost = append(stats.Cost, r.ExecReport.Cost)

	// actionDrop (and not dropped) cause i want the chosen action per class, not the outcome of the request 
	switch r.ClassService.Name {
    case "batch":
        updateActionStats(&stats.Batch, actionDrop, r.ExecReport.SchedAction)
    case "critical-1":
        updateActionStats(&stats.Critical1, actionDrop, r.ExecReport.SchedAction)
    case "critical-2":
        updateActionStats(&stats.Critical2, actionDrop, r.ExecReport.SchedAction)
    default:
        updateActionStats(&stats.Standard, actionDrop, r.ExecReport.SchedAction)
    }

    // actionDrop (and not dropped) cause i want the chosen action per function, not the outcome of the request 
	switch r.Fun.Name {
    case "f1":
        updateActionStats(&stats.f1, actionDrop, r.ExecReport.SchedAction)
    case "f2":
        updateActionStats(&stats.f2, actionDrop, r.ExecReport.SchedAction)
    case "f3":
        updateActionStats(&stats.f3, actionDrop, r.ExecReport.SchedAction)
    case "f4":
        updateActionStats(&stats.f4, actionDrop, r.ExecReport.SchedAction)
    default:
        updateActionStats(&stats.f5, actionDrop, r.ExecReport.SchedAction)
    }

	// 'dropped' cause those are stats for completions
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

	if r.ExecReport.SchedAction == "O_C"{
		if offloadDrop {
			stats.UPCloud[2]++
		} else if outOfTime {
			stats.UPCloud[1]++
		} else {
			stats.UPCloud[0]++
		}
	} else if r.ExecReport.SchedAction == "O_E"{
		if offloadDrop {
			stats.UPEdge[2]++
			// panic("ERRORE (metricGrabberDQN): quì non dovrebbe entrare perchè se sceglie OFFLOADED_EDGE deve poterlo fare!")
		} else if outOfTime {
			stats.UPEdge[1]++
		} else {
			stats.UPEdge[0]++
		}
	} else {
		if actionDrop {
			stats.UPExec[2]++
		} else if outOfTime {
			stats.UPExec[1]++
		} else {
			stats.UPExec[0]++
		}
	}

	switch r.ClassService.Name {
    case "batch":
		if dropped {
			stats.UPBatch[2]++
		} else if outOfTime {
			stats.UPBatch[1]++
		} else {
			stats.UPBatch[0]++
		}
    case "critical-1":
		if dropped {
			stats.UPCritical1[2]++
		} else if outOfTime {
			stats.UPCritical1[1]++
		} else {
			stats.UPCritical1[0]++
		}
    case "critical-2":
		if dropped {
			stats.UPCritical2[2]++
		} else if outOfTime {
			stats.UPCritical2[1]++
		} else {
			stats.UPCritical2[0]++
		}
    default:
		if dropped {
			stats.UPStandard[2]++
		} else if outOfTime {
			stats.UPStandard[1]++
		} else {
			stats.UPStandard[0]++
		}
    }

	switch r.Fun.Name {
    case "f1":
		if dropped {
			stats.UPf1[2]++
		} else if outOfTime {
			stats.UPf1[1]++
		} else {
			stats.UPf1[0]++
		}
    case "f2":
		if dropped {
			stats.UPf2[2]++
		} else if outOfTime {
			stats.UPf2[1]++
		} else {
			stats.UPf2[0]++
		}
    case "f3":
		if dropped {
			stats.UPf3[2]++
		} else if outOfTime {
			stats.UPf3[1]++
		} else {
			stats.UPf3[0]++
		}
    case "f4":
		if dropped {
			stats.UPf4[2]++
		} else if outOfTime {
			stats.UPf4[1]++
		} else {
			stats.UPf4[0]++
		}
    default:
		if dropped {
			stats.UPf5[2]++
		} else if outOfTime {
			stats.UPf5[1]++
		} else {
			stats.UPf5[0]++
		}
    }

    // add stats to InfluxDB every 'updateEvery'
    if elapsedTimeInSeconds - float64(updateEvery*updateRound) > float64(updateEvery) {
    	if elapsedTimeInSeconds-float64(updateEvery*updateRound) > float64(updateEvery) {
	        mg.WriteJSON()
	        stats = EmptyStats()
	        updateRound++
	    }
	}
}


func updateActionStats(slice *[]int, dropped bool, schedAction string) {
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
    // Mappa di nomi delle variabili e i loro valori
    parts := map[string]interface{}{
        "Exec":                	stats.Exec,
        "Cloud":               	stats.Cloud,
        "Edge":                	stats.Edge,
        "Drop":                	stats.Drop,
        "Reward":              	stats.Reward,
        "DeadlinePenalty":     	stats.DeadlinePenalty,
        "DropPenalty":         	stats.DropPenalty,
        "Cost":                	stats.Cost,
        "Standard":            	stats.Standard,
        "Critical1":           	stats.Critical1,
        "Critical2":           	stats.Critical2,
        "Batch":               	stats.Batch,
        "f1":                	stats.f1,
        "f2":            		stats.f2,
        "f3":           		stats.f3,
        "f4":           		stats.f4,
        "f5":               	stats.f5,
        "ResponseTime":        	stats.ResponseTime,
        "IsWarmStart":         	stats.IsWarmStart,
        "InitTime":            	stats.InitTime,
        "Duration":            	stats.Duration,
        "OffloadLatencyCloud":	stats.OffloadLatencyCloud,
        "OffloadLatencyEdge":  	stats.OffloadLatencyEdge,
        "UPExec":             	stats.UPExec,
        "UPCloud":        		stats.UPCloud,
        "UPEdge":         		stats.UPEdge,
        "UPStandard":           stats.UPStandard,
        "UPCritical1":          stats.UPCritical1,
        "UPCritical2": 			stats.UPCritical2,
        "UPBatch":				stats.UPBatch,
        "UPf1":					stats.UPf1,
        "UPf2":					stats.UPf2,
        "UPf3":					stats.UPf3,
        "UPf4":					stats.UPf4,
        "UPf5":					stats.UPf5,
    }

    for name, part := range parts {
        // Convert DQNStats to JSON string
        jsonData, err := json.Marshal(part)
        if err != nil {
            log.Fatalf("%s Error marshalling JSON part '%s': %v\n", INFLUXDB, name, err)
        }

        // Create a new data point
        point := influxdb2.NewPointWithMeasurement("dqn_stats").
            AddTag("name", name).
            AddField("json_data", string(jsonData)).
            SetTime(time.Now().UTC())

        // Write the point to InfluxDB
        err = mg.writeAPI.WritePoint(context.Background(), point)
        if err != nil {
            log.Fatalf("%s Error writing point to InfluxDB for part '%s': %v\n", INFLUXDB, name, err)
        }
    }

    log.Println(INFLUXDB, "Statistics successfully written to InfluxDB")
}


// Closes the InfluxDB client connection
func (mg *metricGrabberDQN) Close() {
	mg.client.Close()
}