package schedule

import (
	"log"
	"time"

	"github.com/tgracchus/assetuploader/pkg/job"
)

// SimpleScheduler is an scheduler for jobs.
type SimpleScheduler interface {
	Schedule(job job.Job) error
}

type immediateScheduler struct {
}

// NewImmediateScheduler creates a Scheduler which inmediately execute jobs.
func NewImmediateScheduler() SimpleScheduler {
	return &immediateScheduler{}
}
func (s *immediateScheduler) Schedule(job job.Job) error {
	return job.Function()
}

// NewSimpleScheduler is a scheduler looking for new jobs every tickPeriod
func NewSimpleScheduler(
	upsert chan job.Job, query chan job.StoreQuery, response chan []job.Job,
	tickPeriod time.Duration,
) SimpleScheduler {
	scheduler := &simpleScheduler{upsert: upsert, query: query, response: response}
	scheduler.tick(tickPeriod)
	return scheduler
}

type simpleScheduler struct {
	upsert   chan job.Job
	query    chan job.StoreQuery
	response chan []job.Job
}

func (s *simpleScheduler) Schedule(scheduledJob job.Job) error {
	// if job is overdued, execute it now
	if time.Now().Before(scheduledJob.ExecutionDate) {
		s.execute(scheduledJob)
		return nil
	}
	return job.UpSert(s.upsert, scheduledJob)
}

func (s *simpleScheduler) execute(scheduledJob job.Job) {
	scheduledJob = scheduledJob.Executing()
	err := job.UpSert(s.upsert, scheduledJob)
	if err != nil {
		log.Println(err.Error())
	}
	err = scheduledJob.Function()
	if err != nil {
		log.Println(err.Error())
		scheduledJob = scheduledJob.Error(err)
		err = job.UpSert(s.upsert, scheduledJob)
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		scheduledJob = scheduledJob.Completed()
		err = job.UpSert(s.upsert, scheduledJob)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (s *simpleScheduler) tick(checkTime time.Duration) {
	ticker := time.NewTicker(checkTime)
	go func() {
		for range ticker.C {
			//TODO: look also for executing overdued jobs
			jobs, err := job.GetBefore(s.query, s.response, time.Now(), []job.Status{job.NewStatus})
			if err != nil {
				log.Println(err.Error())
			}
			for _, job := range jobs {
				s.execute(job)
			}
		}
	}()

}
