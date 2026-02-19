package scheduler

import "time"

// Job represents a task to be executed
type Job struct {
	ID             string
	Payload        string
	CronExpr       string        // Example: "*/5 * * * *"
	NextRun        time.Time     // When this job should run next
	RepeatInterval time.Duration // Interval for recurring jobs
	MaxRuns        int           // Maximum number of runs (0 = unlimited)
	RunCount       int           // Number of times the job has been executed
}

// PriorityQueue implements heap.Interface
type JobQueue []*Job

func (jq JobQueue) Len() int           { return len(jq) }
func (jq JobQueue) Less(i, j int) bool { return jq[i].NextRun.Before(jq[j].NextRun) }
func (jq JobQueue) Swap(i, j int)      { jq[i], jq[j] = jq[j], jq[i] }

func (jq *JobQueue) Push(x interface{}) {
	item := x.(*Job)
	*jq = append(*jq, item)
}

func (jq *JobQueue) Pop() interface{} {
	old := *jq
	n := len(old)
	item := old[n-1]
	*jq = old[0 : n-1]
	return item
}
