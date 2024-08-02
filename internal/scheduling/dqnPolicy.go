package scheduling

import (
	"github.com/grussorusso/serverledge/internal/config"
	"github.com/grussorusso/serverledge/internal/node"
	"log"
)

type DQNPolicy struct {
}

var engine decisionEngine

func (p *DQNPolicy) Init() {
	// initialize decision engine
	version := config.GetString(config.STORAGE_VERSION, "flux")
	if version == "mem" {
		engine = &decisionEngineMem{
			&metricGrabberMem{},
		}
	} else {
		engine = &decisionEngineFlux{
			&metricGrabberFlux{},
		}
	}

	log.Println("Scheduler version:", version)
	engine.InitDecisionEngine()
}

func (p *DQNPolicy) OnCompletion(r *scheduledRequest) {
	//log.Printf("Completed execution of %s in %f\n", r.Fun.Name, r.ExecReport.ResponseTime)
	if r.ExecReport.SchedAction == SCHED_ACTION_OFFLOAD_CLOUD {
		engine.Completed(r, OFFLOADED_CLOUD)
	} else if r.ExecReport.SchedAction == SCHED_ACTION_OFFLOAD_EDGE {
		engine.Completed(r, OFFLOADED_EDGE)
	} else {
		engine.Completed(r, LOCAL)
	}
}

func (p *DQNPolicy) OnArrival(r *scheduledRequest) {
	dec := engine.Decide(r)

	if dec == LOCAL_EXEC_REQUEST {
		containerID, err := node.AcquireWarmContainer(r.Fun)
		if err == nil {
			//log.Printf("Using a warm container for: %v", r)
			execLocally(r, containerID, true)
		} else if handleColdStart(r) {
			//log.Printf("No warm containers for: %v - COLD START", r)
			return
		} else if r.CanDoOffloading {
			if policyFlag == "edgeCloud" {
				// horizontal offloading - search for a nearby node to offload
				//log.Printf("No warm containers and node cant'handle cold start due to lack of resources: proceeding with offloading")
				url := pickEdgeNodeForOffloading(r)
				if url != "" {
					//log.Printf("Found node at url: %s - proceeding with horizontal offloading", url)
					handleEdgeOffload(r, url)
				} else {
					tryCloudOffload(r)
				}
			} else {
				tryCloudOffload(r)
			}
		} else {
			//log.Printf("Can't execute locally and can't offload - dropping incoming request")
			dropRequest(r)
		}
	} else if dec == CLOUD_OFFLOAD_REQUEST {
		// TODO replace handleCloudOffload with tryCloudOffload to apply strict budget enforce
		handleCloudOffload(r)
	} else if dec == EDGE_OFFLOAD_REQUEST {
		url := pickEdgeNodeForOffloading(r)
		if url != "" {
			handleEdgeOffload(r, url)
		} else {
			// log.Println("Can't execute horizontal offloading due to lack of resources available: offloading to cloud")
			tryCloudOffload(r)
		}
	} else if dec == DROP_REQUEST {
		dropRequest(r)
	}
}
