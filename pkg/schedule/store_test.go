package schedule_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tgracchus/assertuploader/pkg/schedule"
)

func TestAddAndGetJob(t *testing.T) {
	jobStore := schedule.NewMemoryJobStore(schedule.MillisBucketKey)
	testJobFunction := func() error {
		return nil
	}
	executionDate := time.Now()
	expectedJob := schedule.NewFixedDateJob(uuid.New().String(), testJobFunction, executionDate)
	err := jobStore.UpSert(*expectedJob)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(250 * time.Millisecond)
	jobs, err := jobStore.GetBefore(executionDate, []schedule.JobStatus{expectedJob.Status})
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
	jobStore := schedule.NewMemoryJobStore(schedule.MillisBucketKey)
	testJobFunction := func() error {
		return nil
	}
	executionDate := time.Now()
	newJob := schedule.NewFixedDateJob(uuid.New().String(), testJobFunction, executionDate)
	err := jobStore.UpSert(*newJob)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(250 * time.Millisecond)
	foundJobs, err := jobStore.GetBefore(executionDate, []schedule.JobStatus{newJob.Status})
	if err != nil {
		t.Fatal(err)
	}
	updatedJob := newJob.Executing()
	err = jobStore.UpSert(updatedJob)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(250 * time.Millisecond)
	updatedFoundJobs, err := jobStore.GetBefore(executionDate, []schedule.JobStatus{updatedJob.Status})
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
	jobStore := schedule.NewMemoryJobStore(schedule.MillisBucketKey)
	testJobFunction := func() error {
		return nil
	}
	now := time.Now()
	pastExecutionDate := now.Add(-1 * time.Hour)
	pastJob := schedule.NewFixedDateJob(uuid.New().String(), testJobFunction, pastExecutionDate)
	err := jobStore.UpSert(*pastJob)
	if err != nil {
		t.Fatal(err)
	}
	newJob := schedule.NewFixedDateJob(uuid.New().String(), testJobFunction, now)
	err = jobStore.UpSert(*newJob)
	if err != nil {
		t.Fatal(err)
	}

	// Since JobStore follows PRAM consistency model,
	// we need to wait for the add channel to be drained, so we can observe the two jobs
	time.Sleep(250 * time.Millisecond)
	foundJobs, err := jobStore.GetBefore(now, []schedule.JobStatus{newJob.Status})
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
