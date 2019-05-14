package schedule_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tgracchus/assertuploader/pkg/job"
	"github.com/tgracchus/assertuploader/pkg/schedule"
)

func TestScheduleJob(t *testing.T) {
	jobs := make(map[string]job.Job)
	simpleScheduler := schedule.NewSimpleScheduler(NewMockJobStore(jobs), 200*time.Millisecond)
	executionDate := time.Now()

	job := job.NewFixedDateJob(uuid.New().String(), jobCallBack, executionDate)
	simpleScheduler.Schedule(*job)
	// Need to wait for the first tick at least
	time.Sleep(500 * time.Millisecond)

	if job, ok := jobs[job.ID]; ok {
		if !job.IsCompleted() {
			t.Fatal("We expect the job to be done")
		}
	} else {
		t.Fatal("We expect the job to be scheduled")
	}
}

func TestScheduleJobFails(t *testing.T) {
	jobs := make(map[string]job.Job)
	simpleScheduler := schedule.NewSimpleScheduler(NewMockJobStore(jobs), 200*time.Millisecond)
	executionDate := time.Now()

	job := job.NewFixedDateJob(uuid.New().String(), errorCallBack, executionDate)
	simpleScheduler.Schedule(*job)
	// Need to wait for the first tick at least
	time.Sleep(500 * time.Millisecond)

	if job, ok := jobs[job.ID]; ok {
		if !job.IsError() {
			t.Fatalf("We expect the job to be failed, not %s", job.Status)
		}
		if job.StatusMsg != "errorCallBack" {
			t.Fatal("We expect the status message to be errorCallBack")
		}
	} else {
		t.Fatal("We expect the job to be scheduled")
	}
}

var jobCallBack = func() error {
	return nil
}

var errorCallBack = func() error {
	return errors.New("errorCallBack")
}

func NewMockJobStore(jobs map[string]job.Job) job.Store {
	return &mockJobStore{jobs: jobs}
}

type mockJobStore struct {
	jobs map[string]job.Job
}

func (ms *mockJobStore) UpSert(job job.Job) error {
	ms.jobs[job.ID] = job
	return nil
}

func (ms *mockJobStore) GetBefore(date time.Time, statuses []job.Status) ([]job.Job, error) {
	jobsList := make([]job.Job, len(ms.jobs))
	i := 0
	for _, v := range ms.jobs {
		jobsList[i] = v
		i++
	}
	return jobsList, nil
}
