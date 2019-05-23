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
	upsert chan job.Job, queries chan job.StoreQuery,
	tickPeriod time.Duration,
) SimpleScheduler {
	scheduler := &simpleScheduler{upsert: upsert, queries: queries, tickPeriod: tickPeriod}
	scheduler.executionLoop()
	return scheduler
}

type simpleScheduler struct {
	upsert     chan job.Job
	queries    chan job.StoreQuery
	tickPeriod time.Duration
}

func (s *simpleScheduler) Schedule(ctx context.Context, scheduledJob job.Job) error {
	return job.UpSert(ctx, s.upsert, scheduledJob)
}

func (s *simpleScheduler) executionLoop() {
	ticker := time.NewTicker(s.tickPeriod)
	go func() {
		for range ticker.C {
			s.executeJobs()
		}
	}()
}

func (s *simpleScheduler) executeJobs() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs, err := job.GetBefore(ctx, s.queries, time.Now(), func(job job.Job) bool {
		overdued := job.ExecutionDate.Add(s.tickPeriod * time.Duration(2))
		return job.IsNew() || (job.IsExecuting() && time.Now().After(overdued))
	})
	if err != nil {
		log.Println(err.Error())
	}
	if jobs != nil {
		for _, jobUnit := range jobs {
			s.executeJob(ctx, jobUnit)
		}
	}
}

func (s *simpleScheduler) executeJob(ctx context.Context, scheduledJob job.Job) {
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
