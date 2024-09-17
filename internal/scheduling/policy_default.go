package scheduling

import (
	"errors"
	"log"

	"github.com/grussorusso/serverledge/internal/config"
	"github.com/grussorusso/serverledge/internal/node"
)

type DefaultLocalPolicy struct {
	queue queue
	mg *metricGrabberDQN
}

func (p *DefaultLocalPolicy) Init() {
	queueCapacity := config.GetInt(config.SCHEDULER_QUEUE_CAPACITY, 0)
	if queueCapacity > 0 {
		log.Printf("Configured queue with capacity %d", queueCapacity)
		p.queue = NewFIFOQueue(queueCapacity)
	} else {
		p.queue = nil
	}
	needStats := config.GetBool(config.DEFAULT_STATS, false)
	if needStats{
		p.mg = InitMG()
	}
}

func (p *DefaultLocalPolicy) OnCompletion(completed *scheduledRequest) {
	if p.queue == nil {
		if p.mg != nil{
			p.mg.addDefaultStats(completed,false)
		}
		return
	}

	p.queue.Lock()
	defer p.queue.Unlock()
	if p.queue.Len() == 0 {
		return
	}

	req := p.queue.Front()

	containerID, err := node.AcquireWarmContainer(req.Fun)
	if err == nil {
		p.queue.Dequeue()
		log.Printf("[%s] Warm start from the queue (length=%d)", req, p.queue.Len())
		execLocally(req, containerID, true)
		return
	}

	if errors.Is(err, node.NoWarmFoundErr) {
		if node.AcquireResources(req.Fun.CPUDemand, req.Fun.MemoryMB, true) {
			log.Printf("[%s] Cold start from the queue", req)
			p.queue.Dequeue()

			// This avoids blocking the thread during the cold
			// start, but also allows us to check for resource
			// availability before dequeueing
			go func() {
				newContainer, err := node.NewContainerWithAcquiredResources(req.Fun)
				if err != nil {
					dropRequest(req)
				} else {
					execLocally(req, newContainer, false)
				}
			}()
			return
		}
	} else if errors.Is(err, node.OutOfResourcesErr) {
	} else {
		// other error
		p.queue.Dequeue()
		dropRequest(req)
	}
}

func (p *DefaultLocalPolicy) OnArrival(r *scheduledRequest) {
	containerID, err := node.AcquireWarmContainer(r.Fun)
	if err == nil {
		execLocally(r, containerID, true)
		return
	}

	if errors.Is(err, node.NoWarmFoundErr) {
		if handleColdStart(r) {
			return
		}
	} else if errors.Is(err, node.OutOfResourcesErr) {
		// pass
	} else {
		// other error
		dropRequest(r)
		if p.mg != nil{
			p.mg.addDefaultStats(r,true)
		}
		return
	}

	// enqueue if possible
	if p.queue != nil {
		p.queue.Lock()
		defer p.queue.Unlock()
		if p.queue.Enqueue(r) {
			log.Printf("[%s] Added to queue (length=%d)", r, p.queue.Len())
			return
		}
	}

	dropRequest(r)
	if p.mg != nil{
		p.mg.addDefaultStats(r,true)
	}
}
