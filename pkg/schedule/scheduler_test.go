package schedule_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tgracchus/assetuploader/pkg/job"
	"github.com/tgracchus/assetuploader/pkg/schedule"
)

func TestScheduleJob(t *testing.T) {
	upsert, query := job.NewMemoryStore(job.MillisKeys)
	simpleScheduler := schedule.NewSimpleScheduler(upsert, query, 200*time.Millisecond)
	executionDate := time.Now()
	ctx := context.Background()
	newJob := job.NewFixedDateJob(uuid.New().String(), jobCallBack, executionDate)
	simpleScheduler.Schedule(ctx, *newJob)
	// Need to wait for the first tick at least
	time.Sleep(500 * time.Millisecond)
	jobs, err := job.GetBefore(ctx, query, time.Now(), newSchedulerTestCriteria(job.CompletedStatus))
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

func TestScheduleJobCancel(t *testing.T) {
	upsert, query := job.NewMemoryStore(job.MillisKeys)
	simpleScheduler := schedule.NewSimpleScheduler(upsert, query, 200*time.Millisecond)
	executionDate := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	newJob := job.NewFixedDateJob(uuid.New().String(), jobCallBack, executionDate)
	simpleScheduler.Schedule(ctx, *newJob)
	// Need to wait for the first tick at least
	time.Sleep(500 * time.Millisecond)
	_, err := job.GetBefore(ctx, query, time.Now(), newSchedulerTestCriteria(job.CompletedStatus))
	if err == nil {
		t.Fatal(err)
	}
	if err.Error() != "context canceled" {
		t.Fatalf("We expect the context to be cancelled")
	}
}
func TestScheduleJobFails(t *testing.T) {
	upsert, query := job.NewMemoryStore(job.MillisKeys)
	simpleScheduler := schedule.NewSimpleScheduler(upsert, query, 200*time.Millisecond)
	executionDate := time.Now()
	ctx := context.Background()
	newJob := job.NewFixedDateJob(uuid.New().String(), errorCallBack, executionDate)
	simpleScheduler.Schedule(ctx, *newJob)
	// Need to wait for the first tick at least
	time.Sleep(500 * time.Millisecond)
	jobs, err := job.GetBefore(ctx, query, time.Now(), newSchedulerTestCriteria(job.ErrorStatus))
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
func TestExecutedOverduedJob(t *testing.T) {
	ctx := context.Background()
	upsert, query := job.NewMemoryStore(job.MillisKeys)
	tick := 100 * time.Millisecond
	simpleScheduler := schedule.NewSimpleScheduler(upsert, query, tick)
	jobExecuted := false
	var wg sync.WaitGroup
	wg.Add(1)
	callback := func(ctx context.Context) error {
		wg.Done()
		return nil
	}
	executionDate := time.Now().Add(tick * -2)
	fixedDateJob := job.NewFixedDateJob(uuid.New().String(), callback, executionDate)
	overduedJob := fixedDateJob.Executing()
	simpleScheduler.Schedule(ctx, overduedJob)
	// Need to wait for the first tick at least
	jobExecuted = waitTimeout(&wg, 2*time.Second)
	if !jobExecuted {
		t.Fatal("We expect the job to be executed")
	}
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return true // completed normally
	case <-time.After(timeout):
		return false // timed out
	}
}

var jobCallBack = func(ctx context.Context) error {
	return nil
}

var errorCallBack = func(ctx context.Context) error {
	return errors.New("errorCallBack")
}

func newSchedulerTestCriteria(status job.Status) func(job job.Job) bool {
	return func(job job.Job) bool {
		return job.Status == status
	}
}
