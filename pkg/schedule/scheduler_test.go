package schedule_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tgracchus/assetuploader/pkg/job"
	"github.com/tgracchus/assetuploader/pkg/schedule"
)

func TestScheduleJob(t *testing.T) {
	upsert, query, responses := job.NewStore(job.NewMemoryStore(job.MillisKeys))
	simpleScheduler := schedule.NewSimpleScheduler(upsert, query, responses, 200*time.Millisecond)
	executionDate := time.Now()

	newJob := job.NewFixedDateJob(uuid.New().String(), jobCallBack, executionDate)
	simpleScheduler.Schedule(*newJob)
	// Need to wait for the first tick at least
	time.Sleep(500 * time.Millisecond)
	jobs, err := job.GetBefore(query, responses, time.Now(), []job.Status{job.CompletedStatus})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("We are expecting 1 job, not %d", len(jobs))
	}
	foundJob := jobs[0]
	if foundJob.ID != newJob.ID {
		t.Fatalf("We recover a different jobs")
	}
}

func TestScheduleJobFails(t *testing.T) {
	upsert, query, responses := job.NewStore(job.NewMemoryStore(job.MillisKeys))
	simpleScheduler := schedule.NewSimpleScheduler(upsert, query, responses, 200*time.Millisecond)
	executionDate := time.Now()

	newJob := job.NewFixedDateJob(uuid.New().String(), errorCallBack, executionDate)
	simpleScheduler.Schedule(*newJob)
	// Need to wait for the first tick at least
	time.Sleep(500 * time.Millisecond)
	jobs, err := job.GetBefore(query, responses, time.Now(), []job.Status{job.ErrorStatus})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("We are expecting 1 job, not %d", len(jobs))
	}
	foundJob := jobs[0]
	if foundJob.ID != newJob.ID {
		t.Fatalf("We recover a different jobs")
	}
	if !foundJob.IsError() {
		t.Fatalf("We expect the job to be failed, not %s", foundJob.Status)
	}
	if foundJob.StatusMsg != "errorCallBack" {
		t.Fatal("We expect the status message to be errorCallBack")
	}

}

var jobCallBack = func() error {
	return nil
}

var errorCallBack = func() error {
	return errors.New("errorCallBack")
}
