package scheduler

import (
	"github.com/google/uuid"
	"time"
)

type JobStatus string

const new JobStatus = "new"
const scheduled JobStatus = "scheduled"
const executing JobStatus = "executing"
const done JobStatus = "done"

func NewFixedDateJob(id uuid.UUID, jobFunction JobFunction, executionDate time.Time) *Job {
	return &Job{id: id, jobFunction: jobFunction, status: new, executionDate: executionDate}
}

type Job struct {
	id            uuid.UUID
	jobFunction   JobFunction
	status        JobStatus
	executionDate time.Time
}

type JobFunction func() error
