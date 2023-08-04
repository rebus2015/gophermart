package agent

import (
	"context"
	"sync"

	"github.com/rebus2015/gophermart/cmd/internal/logger"
)

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, errCh chan<- Result, Cancel chan bool, log *logger.Logger) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			// fan-in job execution multiplexing errCh into the errCh channel
			errCh <- job.execute(ctx)
		case <-ctx.Done():
			log.Info().Msgf("cancelled worker. Error detail: %v\n", ctx.Err())
			errCh <- Result{
				Err: ctx.Err(),
			}
		case <-Cancel:
			log.Info().Msgf("STOP worker")
			return
		}
	}
}

type WorkerPool struct {
	workersCount int
	log          *logger.Logger
	jobs         chan Job
	errCh        chan Result
	Done         chan struct{}
	Cancel       chan bool
	wg           sync.WaitGroup
	workersCtx   context.Context
}

func New(wcount int, lg *logger.Logger) *WorkerPool {
	return &WorkerPool{
		workersCount: wcount,
		jobs:         make(chan Job, wcount),
		errCh:        make(chan Result, wcount),
		Done:         make(chan struct{}),
		Cancel:       make(chan bool, wcount),
		log:          lg,
	}
}

func (wp *WorkerPool) Run(ctx context.Context) {
	for i := 0; i < wp.workersCount; i++ {
		wp.wg.Add(1)
		go worker(wp.workersCtx, &wp.wg, wp.jobs, wp.errCh, wp.Cancel, wp.log)
	}
	wp.wg.Wait()
	close(wp.Done)
	close(wp.errCh)
	close(wp.Cancel)
}

func (wp *WorkerPool) Stop() {
	wp.log.Info().Msgf("Trying to stop workerpool and then start")
	go func() {
		for i := 0; i < wp.workersCount; i++ {
			wp.Cancel <- true
		}
	}()
}

func (wp *WorkerPool) ErrCh() <-chan Result {
	return wp.errCh
}

func (wp *WorkerPool) GenerateFrom(jobsBulk []Job) {
	wp.log.Printf("Generated Jobs channel from %v jobs", len(jobsBulk))
	for i := range jobsBulk {
		wp.jobs <- jobsBulk[i]
	}
	close(wp.jobs)
}
