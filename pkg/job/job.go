package job

import "time"

type Status string

func NewFixedDateJob(id string, jobFunction JobFunction, executionDate time.Time) *Job {
	return &Job{ID: id, JobFunction: jobFunction, Status: NewStatus, StatusMsg: "Job is new", ExecutionDate: executionDate}
}

type Job struct {
	ID            string      `json:"id"`
	JobFunction   JobFunction `json:"-"`
	Status        Status      `json:"status"`
	StatusMsg     string      `json:"statusMsg"`
	ExecutionDate time.Time   `json:"date"`
}

const NewStatus Status = "new"
const ErrorStatus Status = "error"
const ExecutingStatus Status = "executing"
const CompletedStatus Status = "completed"

func (j *Job) IsNew() bool {
	return j.Status == NewStatus
}

func (j *Job) IsCompleted() bool {
	return j.Status == CompletedStatus
}

func (j *Job) IsExecuting() bool {
	return j.Status == ExecutingStatus
}
func (j *Job) IsError() bool {
	return j.Status == ErrorStatus
}

func (j *Job) Completed() Job {
	return j.copy(CompletedStatus, "Job was complete succesfully")
}

func (j *Job) Error(err error) Job {
	return j.copy(ErrorStatus, err.Error())
}

func (j *Job) Executing() Job {
	return j.copy(ExecutingStatus, "Job is being executed")
}

func (j *Job) copy(status Status, statusMsg string) Job {
	return Job{
		ID:            j.ID,
		JobFunction:   j.JobFunction,
		Status:        status,
		ExecutionDate: j.ExecutionDate,
		StatusMsg:     statusMsg,
	}
}

type JobFunction func() error
