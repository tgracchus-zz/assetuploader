package scheduler

import (
	"github.com/tgracchus/assertuploader/pkg/auerrors"
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

func NewSimpleScheduler(Store Store) SimpleScheduler {
	scheduler := &simpleScheduler{store: Store}
	scheduler.start()
	return scheduler
}

type simpleScheduler struct {
	store Store
}

func (s *simpleScheduler) Schedule(job *Job) error {
	err := s.store.Add(job)
	if err != nil {
		return err
	}
	return nil
}

func (s *simpleScheduler) start(){

}

