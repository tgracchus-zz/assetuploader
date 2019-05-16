package schedule

import (
	"context"
	"log"
	"time"

	"github.com/tgracchus/assetuploader/pkg/job"
)

// SimpleScheduler is an scheduler for jobs.
type SimpleScheduler interface {
	Schedule(ctx context.Context, job job.Job) error
}

type immediateScheduler struct {
}

// NewImmediateScheduler creates a Scheduler which inmediately execute jobs.
func NewImmediateScheduler() SimpleScheduler {
	return &immediateScheduler{}
}
func (s *immediateScheduler) Schedule(ctx context.Context, job job.Job) error {
	return job.Function(ctx)
}

// NewSimpleScheduler is a scheduler looking for new jobs every tickPeriod
func NewSimpleScheduler(
	upsert chan job.Job, query chan job.StoreQuery,
	tickPeriod time.Duration,
) SimpleScheduler {
	scheduler := &simpleScheduler{upsert: upsert, query: query}
	scheduler.tick(tickPeriod)
	return scheduler
}

type simpleScheduler struct {
	upsert chan job.Job
	query  chan job.StoreQuery
}

func (s *simpleScheduler) Schedule(ctx context.Context, scheduledJob job.Job) error {
	// if job is overdued, execute it now
	if time.Now().Before(scheduledJob.ExecutionDate) {
		s.execute(ctx, scheduledJob)
		return nil
	}
	return job.UpSert(ctx, s.upsert, scheduledJob)
}

func (s *simpleScheduler) execute(ctx context.Context, scheduledJob job.Job) {
	scheduledJob = scheduledJob.Executing()
	err := job.UpSert(ctx, s.upsert, scheduledJob)
	if err != nil {
		log.Println(err.Error())
	}
	err = scheduledJob.Function(ctx)
	if err != nil {
		log.Println(err.Error())
		scheduledJob = scheduledJob.Error(err)
		err = job.UpSert(ctx, s.upsert, scheduledJob)
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		scheduledJob = scheduledJob.Completed()
		err = job.UpSert(ctx, s.upsert, scheduledJob)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (s *simpleScheduler) tick(checkTime time.Duration) {
	ticker := time.NewTicker(checkTime)
	go func() {
		for range ticker.C {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			jobs, err := job.GetBefore(ctx, s.query, time.Now(), []job.Status{job.NewStatus})
			if err != nil {
				log.Println(err.Error())
			}
			for _, job := range jobs {
				s.execute(ctx, job)

			}
		}
	}()

}
