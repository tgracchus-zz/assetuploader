package job

import "time"

// NewFixedDateJob creates a new job with a fixed execution date.
func NewFixedDateJob(id string, jobFunction Function, executionDate time.Time) *Job {
	return &Job{ID: id, Function: jobFunction, Status: NewStatus, StatusMsg: "Job is new", ExecutionDate: executionDate}
}

// Job represents a job with is id, status, status msg and Function to execute.
type Job struct {
	ID            string    `json:"id"`
	Function      Function  `json:"-"`
	Status        Status    `json:"status"`
	StatusMsg     string    `json:"statusMsg"`
	ExecutionDate time.Time `json:"date"`
}

//Status is the job status type
type Status string

// NewStatus is the status of a new job.
const NewStatus Status = "new"

// ErrorStatus is the status of an errored job.
const ErrorStatus Status = "error"

// ExecutingStatus is the status of an executing job.
const ExecutingStatus Status = "executing"

// CompletedStatus is the status of a completed job.
const CompletedStatus Status = "completed"

// IsNew if the job has the status New.
func (j *Job) IsNew() bool {
	return j.Status == NewStatus
}

// IsCompleted if the job has the status Completed.
func (j *Job) IsCompleted() bool {
	return j.Status == CompletedStatus
}

// IsExecuting if the job has the status Executing.
func (j *Job) IsExecuting() bool {
	return j.Status == ExecutingStatus
}

// IsError if the job has the status Error.
func (j *Job) IsError() bool {
	return j.Status == ErrorStatus
}

// Completed sets the Completed status to a new copy of the job.
func (j *Job) Completed() Job {
	return j.copy(CompletedStatus, "Job was complete succesfully")
}

// Error sets the Error status to a new copy of the job.
func (j *Job) Error(err error) Job {
	return j.copy(ErrorStatus, err.Error())
}

// Executing sets the Executing status to a new copy of the job.
func (j *Job) Executing() Job {
	return j.copy(ExecutingStatus, "Job is being executed")
}

func (j *Job) copy(status Status, statusMsg string) Job {
	return Job{
		ID:            j.ID,
		Function:      j.Function,
		Status:        status,
		ExecutionDate: j.ExecutionDate,
		StatusMsg:     statusMsg,
	}
}

// Function is the function attached to a given job, if it returns and error, the job will marked with an Error status.
type Function func() error
