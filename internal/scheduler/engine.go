package scheduler

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"
)

type Storer interface {
	SaveJob(j *Job) error
	GetJob(id string) (*Job, error)
	GetAllJobs() ([]*Job, error)
	DeleteJob(id string) error
	Close() error
}

type Engine struct {
	queue      JobQueue
	mu         sync.Mutex
	newJobChan chan struct{}
	Storage    Storer
}

func NewEngine() *Engine {
	e := &Engine{
		queue:      make(JobQueue, 0),
		newJobChan: make(chan struct{}, 1),
	}
	heap.Init(&e.queue)
	return e
}

// AddJob adds a new job to the heap and triggers a timer re-evaluation
func (e *Engine) AddJob(j *Job) {
	e.mu.Lock()

	if e.Storage != nil {
		if err := e.Storage.SaveJob(j); err != nil {
			log.Printf("[Error] Failed to save job %s at runtime: %v\n", j.ID, err)
		}
	}

	heap.Push(&e.queue, j)
	e.mu.Unlock()

	// Notify engine of new job (NextRun might be earlier than currently waited)
	select {
	case e.newJobChan <- struct{}{}:
	default:
	}
}

func (e *Engine) Run(ctx context.Context) {
	log.Println("Scheduler Engine started...")

	for {
		e.mu.Lock()
		var nextRun time.Duration = 1 * time.Hour // Default wait if queue is empty

		if e.queue.Len() > 0 {
			now := time.Now()
			nextJob := e.queue[0]

			if now.After(nextJob.NextRun) || now.Equal(nextJob.NextRun) {
				// Time to execute!
				job := heap.Pop(&e.queue).(*Job)
				e.mu.Unlock()

				go e.execute(job)
				continue
			}
			nextRun = time.Until(nextJob.NextRun)
		}
		e.mu.Unlock()

		timer := time.NewTimer(nextRun)

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			// Timer expired, loop again to execute top job
		case <-e.newJobChan:
			// New job received, reset timer to check if this new job is more urgent
			timer.Stop()
		}
	}
}

func (e *Engine) execute(j *Job) {
	// Execution must be safe from panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Recover] Job %s failed: %v\n", j.ID, r)
		}
	}()

	log.Printf("[%s] Executing job: %s\n", time.Now().Format("15:04:05"), j.Payload)
	// Simulate work
	time.Sleep(100 * time.Millisecond)

	// Delete job from storage to prevent re-execution on restart
	if e.Storage != nil {
		if err := e.Storage.DeleteJob(j.ID); err != nil {
			log.Printf("[Error] Failed to delete job %s: %v\n", j.ID, err)
		} else {
			log.Printf("[Storage] Job %s deleted from disk\n", j.ID)
		}
	}
}
