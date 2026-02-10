package scheduler

import "time"

// Job representasi tugas yang akan dieksekusi
type Job struct {
	ID       string
	Payload  string
	CronExpr string    // Misal: "*/5 * * * *"
	NextRun  time.Time // Kapan job ini harus jalan selanjutnya
}

// PriorityQueue akan mengimplementasikan heap.Interface
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
