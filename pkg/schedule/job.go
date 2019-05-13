package schedule

import "time"

type JobStatus string

func NewFixedDateJob(id string, jobFunction JobFunction, executionDate time.Time) *Job {
	return &Job{ID: id, JobFunction: jobFunction, Status: newJobStatus, ExecutionDate: executionDate}
}

type Job struct {
	ID            string      `json:"id"`
	JobFunction   JobFunction `json:"-"`
	Status        JobStatus   `json:"status"`
	StatusMsg     string      `json:"statusMsg"`
	ExecutionDate time.Time   `json:"date"`
}

const newJobStatus JobStatus = "new"
const errorJobStatus JobStatus = "error"
const executingJobStatus JobStatus = "executing"
const completedJobStatus JobStatus = "completed"

func (j *Job) IsNew() bool {
	return j.Status == newJobStatus
}

func (j *Job) IsCompleted() bool {
	return j.Status == completedJobStatus
}

func (j *Job) IsExecuting() bool {
	return j.Status == executingJobStatus
}
func (j *Job) IsError() bool {
	return j.Status == errorJobStatus
}

func (j *Job) Completed() Job {
	return j.copy(completedJobStatus, "Job was complete succesfully")
}

func (j *Job) Error(err error) Job {
	return j.copy(errorJobStatus, err.Error())
}

func (j *Job) Executing() Job {
	return j.copy(executingJobStatus, "Job is being executed")
}

func (j *Job) copy(status JobStatus, statusMsg string) Job {
	return Job{
		ID:            j.ID,
		JobFunction:   j.JobFunction,
		Status:        status,
		ExecutionDate: j.ExecutionDate,
		StatusMsg:     statusMsg,
	}
}

type JobFunction func() error
