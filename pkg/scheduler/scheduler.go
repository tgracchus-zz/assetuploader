package scheduler

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tgracchus/assertuploader/pkg/auerrors"
)

type JobStatus string

func NewFixedDateJob(id uuid.UUID, jobFunction JobFunction, executionDate time.Time) *Job {
	return &Job{ID: id, JobFunction: jobFunction, Status: newJobStatus, ExecutionDate: executionDate}
}

type Job struct {
	ID            uuid.UUID
	JobFunction   JobFunction
	Status        JobStatus
	ExecutionDate time.Time
}

const newJobStatus JobStatus = "new"
const executingJobStatus JobStatus = "executing"
const doneJobStatus JobStatus = "done"

func (j *Job) IsNew() bool {
	return j.Status == newJobStatus
}

func (j *Job) IsDone() bool {
	return j.Status == doneJobStatus
}

func (j *Job) IsExecuting() bool {
	return j.Status == executingJobStatus
}

func (j *Job) copy(status JobStatus) *Job {
	return &Job{
		ID:            j.ID,
		JobFunction:   j.JobFunction,
		Status:        status,
		ExecutionDate: j.ExecutionDate,
	}
}

func (j *Job) Done() *Job {
	return j.copy(doneJobStatus)
}

func (j *Job) Executing() *Job {
	return j.copy(executingJobStatus)
}

type JobFunction func() error

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
		return auerrors.FError(auerrors.ErrorConflict, "Executed before execution date %s", job.ExecutionDate.String())
	}

	return job.JobFunction()
}

func NewSimpleScheduler(Store JobStore, checkTime time.Duration) SimpleScheduler {
	scheduler := &simpleScheduler{store: Store}
	scheduler.start(checkTime)
	return scheduler
}

type simpleScheduler struct {
	store JobStore
}

func (s *simpleScheduler) Schedule(job Job) error {
	err := s.store.UpSert(job)
	if err != nil {
		return err
	}
	return nil
}

func (s *simpleScheduler) start(checkTime time.Duration) {
	ticker := time.NewTicker(checkTime)
	go func() {
		for range ticker.C {
			jobs, err := s.store.GetBefore(time.Now(), []JobStatus{newJobStatus, executingJobStatus})
			if err != nil {
				log.Fatal(err)
			}

			for _, job := range jobs {
				err = job.JobFunction()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

}
