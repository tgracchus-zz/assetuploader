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
func NewSimpleScheduler(Store job.Store, tickPeriod time.Duration) SimpleScheduler {
	scheduler := &simpleScheduler{store: Store}
	scheduler.tick(tickPeriod)
	return scheduler
}

type simpleScheduler struct {
	store job.Store
}

func (s *simpleScheduler) Schedule(job job.Job) error {
	// if job is overdued, execute it now
	if time.Now().Before(job.ExecutionDate) {
		s.execute(job)
		return nil
	}
	return s.store.UpSert(job)
}

func (s *simpleScheduler) execute(job job.Job) {
	job = job.Executing()
	err := s.store.UpSert(job)
	if err != nil {
		log.Println(err.Error())
	}
	err = job.Function()
	if err != nil {
		log.Println(err.Error())
		job = job.Error(err)
		err = s.store.UpSert(job)
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		job = job.Completed()
		err = s.store.UpSert(job)
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
			jobs, err := s.store.GetBefore(time.Now(), []job.Status{job.NewStatus})
			if err != nil {
				log.Println(err.Error())
			}
			for _, job := range jobs {
				s.execute(job)
			}
		}
	}()

}
