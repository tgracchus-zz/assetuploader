package scheduler_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tgracchus/assertuploader/pkg/scheduler"
)

func TestAddAndGetJob(t *testing.T) {
	jobStore := scheduler.NewMemoryJobStore(scheduler.MillisBucketKey)
	testJobFunction := func() error {
		return nil
	}
	executionDate := time.Now()
	expectedJob := scheduler.NewFixedDateJob(uuid.New().String(), testJobFunction, executionDate)
	err := jobStore.UpSert(*expectedJob)
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := jobStore.GetBefore(executionDate, []scheduler.JobStatus{expectedJob.Status})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatal("Expected at least one job")
	}
	if jobs[0].ID != expectedJob.ID {
		t.Fatal("Expected job and actual job do not match")
	}
}

func TestUpdateJobStatus(t *testing.T) {
	jobStore := scheduler.NewMemoryJobStore(scheduler.MillisBucketKey)
	testJobFunction := func() error {
		return nil
	}
	executionDate := time.Now()
	newJob := scheduler.NewFixedDateJob(uuid.New().String(), testJobFunction, executionDate)
	err := jobStore.UpSert(*newJob)
	if err != nil {
		t.Fatal(err)
	}
	foundJobs, err := jobStore.GetBefore(executionDate, []scheduler.JobStatus{newJob.Status})
	if err != nil {
		t.Fatal(err)
	}
	updatedJob := newJob.Executing()
	err = jobStore.UpSert(*updatedJob)
	if err != nil {
		t.Fatal(err)
	}
	updatedFoundJobs, err := jobStore.GetBefore(executionDate, []scheduler.JobStatus{updatedJob.Status})
	if err != nil {
		t.Fatal(err)
	}
	if len(updatedFoundJobs) != 1 {
		t.Fatal("Expected at least one job")
	}
	if len(foundJobs) != 1 {
		t.Fatal("Expected at least one job")
	}
	foundJob := foundJobs[0]
	if foundJob.ID != newJob.ID {
		t.Fatal("Expected job and actual job do not match")
	}
	if !newJob.IsNew() {
		t.Fatal("Expected job should be in scheduled state")
	}
	if !foundJob.IsNew() {
		t.Fatal("Expected job should be in scheduled state")
	}
	updatedFoundJob := updatedFoundJobs[0]
	if updatedFoundJob.ID != newJob.ID {
		t.Fatal("Expected job and actual job do not match")
	}
	if updatedFoundJob.ID != updatedJob.ID {
		t.Fatal("Expected job and actual job do not match")
	}
	if !newJob.IsNew() {
		t.Fatal("Expected job should be in scheduled state")
	}
	if !updatedJob.IsExecuting() {
		t.Fatal("Expected job should be in scheduled state")
	}
	if !updatedFoundJob.IsExecuting() {
		t.Fatal("Expected job should be in scheduled state")
	}
}

func TestAddJobPastInTime(t *testing.T) {
	jobStore := scheduler.NewMemoryJobStore(scheduler.MillisBucketKey)
	testJobFunction := func() error {
		return nil
	}
	now := time.Now()
	pastExecutionDate := now.Add(-1 * time.Hour)
	pastJob := scheduler.NewFixedDateJob(uuid.New().String(), testJobFunction, pastExecutionDate)
	err := jobStore.UpSert(*pastJob)
	if err != nil {
		t.Fatal(err)
	}
	newJob := scheduler.NewFixedDateJob(uuid.New().String(), testJobFunction, now)
	err = jobStore.UpSert(*newJob)
	if err != nil {
		t.Fatal(err)
	}
	foundJobs, err := jobStore.GetBefore(now, []scheduler.JobStatus{newJob.Status})
	if err != nil {
		t.Fatal(err)
	}
	if len(foundJobs) != 2 {
		t.Fatal("Expected at least two jobs")
	}
	foundJob := foundJobs[0]
	if newJob.ID != foundJob.ID {
		t.Fatal("Expected job and actual job do not match")
	}
}
