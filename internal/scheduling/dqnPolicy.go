package scheduling

import (
	"github.com/grussorusso/serverledge/internal/node"
	"log"
)

type DQNPolicy struct {
}

var engine decisionEngine

func (p *DQNPolicy) Init() {
	// initialize decision engine
	engine = &decisionEngineDQN{}
	engine.InitDecisionEngine()
}

func (p *DQNPolicy) OnCompletion(r *scheduledRequest) {
	// log.Printf("Completed execution of %s in %f\n", r.Fun.Name, r.ExecReport.ResponseTime)
	if r.ExecReport.SchedAction == SCHED_ACTION_OFFLOAD_CLOUD {
		engine.Completed(r, OFFLOADED_CLOUD)
	} else if r.ExecReport.SchedAction == SCHED_ACTION_OFFLOAD_EDGE {
		engine.Completed(r, OFFLOADED_EDGE)
	} else {
		engine.Completed(r, LOCAL)
	}
}

func (p *DQNPolicy) OnArrival(r *scheduledRequest) {
	dec, urlEdgeNode := engine.Decide(r)

	if dec == LOCAL_EXEC_REQUEST {
		containerID, err := node.AcquireWarmContainer(r.Fun)
		if err == nil {
			// log.Printf("Using a warm container for: %v", r)
			execLocally(r, containerID, true)
		} else if handleColdStart(r) {
			// log.Printf("No warm containers for: %v - COLD START", r)
			return
		/*
		} else if r.CanDoOffloading {
			// horizontal offloading - search for a nearby node to offload
			// log.Printf("No warm containers and node cant'handle cold start due to lack of resources: proceeding with offloading")
			url := pickEdgeNodeForOffloading(r)
			if url != "" {
				// log.Printf("Found node at url: %s - proceeding with horizontal offloading", url)
				handleEdgeOffload(r, url)
			} else {
				tryCloudOffload(r)
			}
		*/
		} else {
			// log.Printf("Can't execute locally and can't offload - dropping incoming request")
			message = " ### ERRORE (dqnPolicy): quì non dovrebbe entrare perchè il filtro dovrebbe impedirglielo!"
			log.Printf(message)
			panic(message)
			// dropRequest(r)
		}
	} else if dec == CLOUD_OFFLOAD_REQUEST {
		handleCloudOffload(r)
	} else if dec == EDGE_OFFLOAD_REQUEST {
		if urlEdgeNode == "" {
			message = " ### ERRORE (dqnPolicy): quì non dovrebbe entrare perchè se sceglie OFFLOADED_EDGE deve poterlo fare!"
			log.Printf(message)
			panic(message)
		}
		handleEdgeOffload(r, urlEdgeNode)
	} else if dec == DROP_REQUEST {
		dropRequest(r)
	}
}
