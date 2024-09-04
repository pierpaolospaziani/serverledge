package scheduling

import (
	"github.com/grussorusso/serverledge/internal/node"
	"log"
)

type ProbabilisticPolicy struct {}

var dEngine decisionEngine

func (p *ProbabilisticPolicy) Init() {
	// initialize decision engine
	dEngine = &decisionEngineProbabilistic{}
	dEngine.InitDecisionEngine()
}

func (p *ProbabilisticPolicy) OnCompletion(r *scheduledRequest) {
	if r.ExecReport.HasBeenDropped {
		dEngine.Completed(r, 1)
	} else {
		dEngine.Completed(r, 0)
	}
}

func (p *ProbabilisticPolicy) OnArrival(r *scheduledRequest) {
	dec := dEngine.Decide(r)

	if dec == LOCAL_EXEC_REQUEST {
		containerID, err := node.AcquireWarmContainer(r.Fun)
		if err == nil {
			log.Printf("Using a warm container for: %v", r)
			execLocally(r, containerID, true)
		} else if handleColdStart(r) {
			log.Printf("No warm containers for: %v - COLD START", r)
			return
		} else if r.CanDoOffloading {
			log.Printf("No warm containers and node cant'handle cold start due to lack of resources: proceeding with offloading")
			url := pickEdgeNodeForOffloading(r)
			if url != "" {
				log.Printf("Found node at url: %s - proceeding with horizontal offloading", url)
				handleEdgeOffload(r, url)
			} else {
				handleCloudOffload(r)
			}
		} else {
			log.Printf("Can't execute locally and can't offload - dropping incoming request")
			dropRequest(r)
		}
	} else if dec == CLOUD_OFFLOAD_REQUEST {
		handleCloudOffload(r)
	} else if dec == EDGE_OFFLOAD_REQUEST {
		url := pickEdgeNodeForOffloading(r)
		if url != "" {
			handleEdgeOffload(r, url)
		} else {
			log.Println("Can't execute horizontal offloading due to lack of resources available: offloading to cloud")
			handleCloudOffload(r)
		}
	} else if dec == DROP_REQUEST {
		dropRequest(r)
	}
}
