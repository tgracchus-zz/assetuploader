package schedule

import (
	"log"
	"time"

	"github.com/tgracchus/assetuploader/pkg/auerr"
	"github.com/tgracchus/assetuploader/pkg/job"
)

type SimpleScheduler interface {
	Schedule(job job.Job) error
}

type immediateScheduler struct {
}

func NewImmediateScheduler() SimpleScheduler {
	return &immediateScheduler{}
}
func (s *immediateScheduler) Schedule(job job.Job) error {
	if time.Now().Before(job.ExecutionDate) {
		return auerr.FError(auerr.ErrorConflict, "Executed before execution date %s", job.ExecutionDate.String())
	}
	return job.Function()
}

func NewSimpleScheduler(Store job.Store, tickPeriod time.Duration) SimpleScheduler {
	scheduler := &simpleScheduler{store: Store}
	scheduler.tick(tickPeriod)
	return scheduler
}

type simpleScheduler struct {
	store job.Store
}

func (s *simpleScheduler) Schedule(job job.Job) error {
	return s.store.UpSert(job)
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
				job = job.Executing()
				err = s.store.UpSert(job)
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
		}
	}()

}
