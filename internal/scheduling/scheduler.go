package scheduling

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/grussorusso/serverledge/internal/metrics"
	"github.com/grussorusso/serverledge/internal/node"

	"github.com/grussorusso/serverledge/internal/config"

	"github.com/grussorusso/serverledge/internal/container"
	"github.com/grussorusso/serverledge/internal/function"
)

var requests chan *scheduledRequest
var completions chan *completion

var remoteServerUrl string
var executionLogEnabled bool

var offloadingClient *http.Client
var policy Policy

func Run(p Policy) {
	policy = p
	requests = make(chan *scheduledRequest, 500)
	completions = make(chan *completion, 500)

	// initialize Resources resources
	availableCores := runtime.NumCPU()
	node.Resources.AvailableMemMB = int64(config.GetInt(config.POOL_MEMORY_MB, 1024))
	node.Resources.AvailableCPUs = config.GetFloat(config.POOL_CPUS, float64(availableCores))
	node.Resources.MaxMemMB = node.Resources.AvailableMemMB
	node.Resources.MaxCPUs = node.Resources.AvailableCPUs
	node.Resources.ContainerPools = make(map[string]*node.ContainerPool)
	// FIXME AUDIT log.Printf("Current resources: %v", node.Resources)

	container.InitDockerContainerFactory()

	//janitor periodically remove expired warm container
	node.GetJanitorInstance()

	tr := &http.Transport{
		MaxIdleConns:        2500,
		MaxIdleConnsPerHost: 2500,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     30 * time.Minute,
	}
	offloadingClient = &http.Client{Transport: tr}

	// initialize scheduling policy
	p.Init()

	remoteServerUrl = config.GetString(config.CLOUD_URL, "")

	log.Println("Scheduler started.")

	var r *scheduledRequest
	var c *completion
	for {
		select {
		case r = <-requests:
			//TODO correct place?
			if metrics.Enabled {
				metrics.AddArrivals(r.Fun.Name, r.ClassService.Name)
			}

			go p.OnArrival(r)
		case c = <-completions:
			// if is Drop from Offload don't do ReleaseContainer
			if !c.scheduledRequest.ExecReport.HasBeenDropped {
				node.ReleaseContainer(c.contID, c.Fun)
			}
			p.OnCompletion(c.scheduledRequest)

			// if is Drop from Offload don't do
			if !c.scheduledRequest.ExecReport.HasBeenDropped {
				// fixme: why always true?
				if metrics.Enabled && (r.ExecReport.SchedAction != SCHED_ACTION_OFFLOAD_CLOUD || r.ExecReport.SchedAction != SCHED_ACTION_OFFLOAD_EDGE) {
					addCompletedMetrics(r)
				}
			}
		}
	}

}

func addCompletedMetrics(r *scheduledRequest) {
	metrics.AddCompletedInvocation(r.Fun.Name)
	metrics.AddFunctionDurationValue(r.Fun.Name, r.ExecReport.Duration)

	if !r.ExecReport.IsWarmStart {
		metrics.AddColdStart(r.Fun.Name, r.ExecReport.InitTime)
	}

}

// SubmitRequest submits a newly arrived request for scheduling and execution
func SubmitRequest(r *function.Request) error {
	schedRequest := scheduledRequest{
		Request:         r,
		decisionChannel: make(chan schedDecision, 1)}
	requests <- &schedRequest

	// wait on channel for scheduling action
	schedDecision, ok := <-schedRequest.decisionChannel
	if !ok {
		return fmt.Errorf("could not schedule the request")
	}
	// FIXME AUDIT log.Printf("[%s] Scheduling decision: %v", r, schedDecision)

	var err error
	if schedDecision.action == DROP {
		// FIXME AUDIT log.Printf("[%s] Dropping request", r)
		return node.OutOfResourcesErr
	} else if schedDecision.action == EXEC_REMOTE || schedDecision.action == EXEC_NEIGHBOUR {
		// FIXME AUDIT log.Printf("Offloading request")
		err = Offload(r, schedDecision.remoteHost)
		if err != nil {
			_, isDQN := policy.(*DQNPolicy)
			_, isProbabilistic := policy.(*ProbabilisticPolicy)
			if (isDQN || isProbabilistic) && err == node.OutOfResourcesErr {
				if checkIfCloudOffloading(schedDecision.remoteHost) {
					r.ExecReport.SchedAction = SCHED_ACTION_OFFLOAD_CLOUD
				} else {
					r.ExecReport.SchedAction = SCHED_ACTION_OFFLOAD_EDGE
				}
				r.ExecReport.HasBeenDropped = true
				completions <- &completion{scheduledRequest: &scheduledRequest{Request: r}}
			}
			return err
		}
	} else {
		err = Execute(schedDecision.contID, &schedRequest)
		if err != nil {
			return err
		}
	}
	return nil
}

// SubmitAsyncRequest submits a newly arrived async request for scheduling and execution
func SubmitAsyncRequest(r *function.Request) {
	schedRequest := scheduledRequest{
		Request:         r,
		decisionChannel: make(chan schedDecision, 1)}
	requests <- &schedRequest

	// wait on channel for scheduling action
	schedDecision, ok := <-schedRequest.decisionChannel
	if !ok {
		publishAsyncResponse(r.ReqId, function.Response{Success: false})
		return
	}

	var err error
	if schedDecision.action == DROP {
		publishAsyncResponse(r.ReqId, function.Response{Success: false})
	} else if schedDecision.action == EXEC_REMOTE || schedDecision.action == EXEC_NEIGHBOUR {
		// FIXME AUDIT log.Printf("Offloading request")
		err = OffloadAsync(r, schedDecision.remoteHost)
		if err != nil {
			publishAsyncResponse(r.ReqId, function.Response{Success: false})
		}
	} else {
		err = Execute(schedDecision.contID, &schedRequest)
		if err != nil {
			publishAsyncResponse(r.ReqId, function.Response{Success: false})
		}
		publishAsyncResponse(r.ReqId, function.Response{Success: true, ExecutionReport: r.ExecReport})
	}
}

func handleColdStart(r *scheduledRequest) (isSuccess bool) {
	newContainer, err := node.NewContainer(r.Fun)
	if errors.Is(err, node.OutOfResourcesErr) {
		return false
	} else if err != nil {
		log.Printf("Cold start failed: %v", err)
		return false
	} else {
		execLocally(r, newContainer, false)
		return true
	}
}

func dropRequest(r *scheduledRequest) {
	r.decisionChannel <- schedDecision{action: DROP}
}

func execLocally(r *scheduledRequest, c container.ContainerID, warmStart bool) {
	initTime := time.Now().Sub(r.Arrival).Seconds()
	r.ExecReport.InitTime = initTime
	r.ExecReport.IsWarmStart = warmStart

	decision := schedDecision{action: EXEC_LOCAL, contID: c}
	r.decisionChannel <- decision
}

func handleOffload(r *scheduledRequest, serverHost string, act action) {
	r.CanDoOffloading = false // the next server can't offload this request
	r.decisionChannel <- schedDecision{
		action:     act,
		contID:     "",
		remoteHost: serverHost,
	}
}

func tryCloudOffload(r *scheduledRequest) {
	if canAffordCloudOffloading(r) {
		handleCloudOffload(r)
	} else {
		dropRequest(r)
	}
}

func handleCloudOffload(r *scheduledRequest) {
	// FIXME MAYBE USELESS cloudAddress := config.GetString(config.CLOUD_URL, "")
	cloudAddress := pickCloudNodeForOffloading()
	// FIXME AUDIT log.Printf("Handling offload to cloud address %s", cloudAddress)
	handleOffload(r, cloudAddress, EXEC_REMOTE)
}

func handleEdgeOffload(r *scheduledRequest, serverHost string) {
	// FIXME AUDIT log.Printf("Handling offload to nearby host %s", serverHost)
	handleOffload(r, serverHost, EXEC_NEIGHBOUR)
}
