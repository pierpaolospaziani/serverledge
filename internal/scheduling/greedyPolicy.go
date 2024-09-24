package scheduling

import (
	"github.com/grussorusso/serverledge/internal/node"
	"log"
)

type GreedyPolicy struct {}

var dEngineGreedy decisionEngine

func (p *GreedyPolicy) Init() {
	// initialize decision engine
	dEngineGreedy = &decisionEngineProbabilistic{}
	dEngineGreedy.InitDecisionEngine()
}

func (p *GreedyPolicy) OnCompletion(r *scheduledRequest) {
	if r.ExecReport.HasBeenDropped {
		dEngineGreedy.Completed(r, 1)
	} else {
		dEngineGreedy.Completed(r, 0)
	}
}

func (p *GreedyPolicy) OnArrival(r *scheduledRequest) {
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
}
