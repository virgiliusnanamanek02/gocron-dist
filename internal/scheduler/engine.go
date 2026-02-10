package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

type Engine struct {
	queue      JobQueue
	mu         sync.Mutex
	newJobChan chan struct{}
}

func NewEngine() *Engine {
	e := &Engine{
		queue:      make(JobQueue, 0),
		newJobChan: make(chan struct{}, 1),
	}
	heap.Init(&e.queue)
	return e
}

// AddJob memasukkan job baru ke heap dan mentrigger evaluasi ulang timer
func (e *Engine) AddJob(j *Job) {
	e.mu.Lock()
	heap.Push(&e.queue, j)
	e.mu.Unlock()

	// Beritahu engine ada job baru (mungkin NextRun-nya lebih awal dari yang sekarang ditunggu)
	select {
	case e.newJobChan <- struct{}{}:
	default:
	}
}

func (e *Engine) Run(ctx context.Context) {
	fmt.Println("Scheduler Engine started...")

	for {
		e.mu.Lock()
		var nextRun time.Duration = 1 * time.Hour // Default wait jika queue kosong

		if e.queue.Len() > 0 {
			now := time.Now()
			nextJob := e.queue[0]

			if now.After(nextJob.NextRun) || now.Equal(nextJob.NextRun) {
				// Waktunya eksekusi!
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
			// Timer habis, waktunya loop lagi untuk eksekusi job teratas
		case <-e.newJobChan:
			// Ada job baru masuk, reset timer untuk cek apakah job baru ini lebih urgen
			timer.Stop()
		}
	}
}

func (e *Engine) execute(j *Job) {
	// Di level senior, eksekusi harus aman dari panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[Recover] Job %s failed: %v\n", j.ID, r)
		}
	}()

	fmt.Printf("[%s] Executing job: %s\n", time.Now().Format("15:04:05"), j.Payload)
	// Simulasi kerja
	time.Sleep(100 * time.Millisecond)
}
