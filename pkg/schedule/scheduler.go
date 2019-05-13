package schedule

import (
	"log"
	"time"

	"github.com/tgracchus/assertuploader/pkg/auerr"
)

type SimpleScheduler interface {
	Schedule(job Job) error
}

type immediateScheduler struct {
}

func NewImmediateScheduler() SimpleScheduler {
	return &immediateScheduler{}
}
func (s *immediateScheduler) Schedule(job Job) error {
	if time.Now().Before(job.ExecutionDate) {
		return auerr.FError(auerr.ErrorConflict, "Executed before execution date %s", job.ExecutionDate.String())
	}
	return job.JobFunction()
}

func NewSimpleScheduler(Store JobStore, tickPeriod time.Duration) SimpleScheduler {
	scheduler := &simpleScheduler{store: Store}
	scheduler.tick(tickPeriod)
	return scheduler
}

type simpleScheduler struct {
	store JobStore
}

func (s *simpleScheduler) Schedule(job Job) error {
	return s.store.UpSert(job)
}

func (s *simpleScheduler) tick(checkTime time.Duration) {
	ticker := time.NewTicker(checkTime)
	go func() {
		for range ticker.C {
			jobs, err := s.store.GetBefore(time.Now(), []JobStatus{newJobStatus})
			if err != nil {
				log.Println(err.Error())
			}
			for _, job := range jobs {
				job = job.Executing()
				err = s.store.UpSert(job)
				if err != nil {
					log.Println(err.Error())
				}
				err = job.JobFunction()
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
