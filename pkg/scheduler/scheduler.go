package scheduler

import (
	"github.com/tgracchus/assertuploader/pkg/auerrors"
	"log"
	"time"
)

type SimpleScheduler interface {
	Schedule(job *Job) error
}

type immediateScheduler struct {
}

func NewImmediateScheduler() SimpleScheduler {
	return &immediateScheduler{}
}
func (s *immediateScheduler) Schedule(job *Job) error {
	if time.Now().Before(job.executionDate) {
		return auerrors.FError(auerrors.ErrorConflict, "Executed before execution date %s", job.executionDate.String())
	}

	return job.jobFunction()
}

func NewSimpleScheduler(Store JobStore, checkTime time.Duration) SimpleScheduler {
	scheduler := &simpleScheduler{store: Store}
	scheduler.start(checkTime)
	return scheduler
}

type simpleScheduler struct {
	store JobStore
}

func (s *simpleScheduler) Schedule(job *Job) error {
	err := s.store.Add(job)
	if err != nil {
		return err
	}
	return nil
}

func (s *simpleScheduler) start(checkTime time.Duration) {
	ticker := time.NewTicker(checkTime)
	go func() {
		for range ticker.C {
			jobs, err := s.store.GetBefore(time.Now())
			if err != nil {
				log.Fatal(err)
			}

			for _, job := range jobs {
				err = job.jobFunction()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

}
