package scheduler

import (
	"github.com/google/uuid"
	"time"
)

type JobStore interface {
	Add(job *Job) error
	GetBefore(date time.Time) ([]Job, error)
}

func NewMemoryJobStore(bucketKeyFunc bucketKeyFunc) JobStore {
	add := make(chan Job)
	query := make(chan time.Time)
	out := make(chan []Job)
	store := &memoryJobStore{add: add, query: query, out: out}
	memoryJobStoreMonitor(bucketKeyFunc,add, query, out)
	return store
}

type memoryJobStore struct {
	add   chan<- Job
	query chan<- time.Time
	out   <-chan []Job
}

func (st *memoryJobStore) Add(job *Job) error {
	st.add <- *job
	return nil
}

func (st *memoryJobStore) GetBefore(date time.Time) ([]Job, error) {
	st.query <- date
	return <-st.out, nil
}

func memoryJobStoreMonitor(bucketKeyFunc bucketKeyFunc, add chan Job, query chan time.Time, out chan []Job) {
	jobs := newTimeBuckets(bucketKeyFunc)
	go func() {
		for {
			select {
			case job, ok := <-add:
				if !ok {
					add = nil
				}
				jobs.add(job)
			case date, ok := <-query:
				if !ok {
					query = nil
				}
				out <- jobs.getBefore(date)
			}

			if add == nil && query == nil {
				break
			}
		}
	}()
}

type bucketKeyFunc func(date time.Time) time.Time

func newTimeBuckets(bucketKeyFunc bucketKeyFunc) *jobs {
	now := bucketKeyFunc(time.Now())
	bucket := newTimeBucket(now, nil)
	buckets := make(map[time.Time]*timeBucket)
	buckets[now] = bucket
	return &jobs{buckets, bucket, bucketKeyFunc}
}

type jobs struct {
	buckets       map[time.Time]*timeBucket
	headBucket    *timeBucket
	bucketKeyFunc bucketKeyFunc
}

func (j *jobs) add(job Job) {
	now := j.bucketKeyFunc(job.executionDate)
	bucket := j.findOrCreateBucket(now)
	bucket.jobs[job.id] = &job
}

func (j *jobs) getBefore(date time.Time) []Job {
	now := j.bucketKeyFunc(date)
	return j.findBucketsBefore(now)

}

func (j *jobs) findBucketsBefore(bucketKey time.Time) []Job {
	bucket := j.headBucket
	jobs := make([]Job, 0, 0)
	for bucket != nil {
		if bucket.bucketKey.Before(bucketKey) {
			for _, job := range bucket.jobs {
				jobs = append(jobs, *job)
			}
		}
		bucket = bucket.previous
	}

	return jobs
}

func (j *jobs) findOrCreateBucket(bucketKey time.Time) *timeBucket {
	bucket := j.headBucket
	var lastBucket *timeBucket = nil
	for bucket != nil {
		//Bucket has the same key than current bucket, just return it
		if bucket.bucketKey.Equal(bucketKey) {
			return bucket
		}
		//Bucket is newer than the current bucket, create it
		if bucketKey.After(bucket.bucketKey) {
			newBucket := newTimeBucket(bucketKey, bucket)
			newBucket.previous = bucket
			//if previous bucket is nil it means we need to change the headBucket
			if lastBucket == nil {
				j.headBucket = newBucket
			}
			return newBucket
		}

		// Just iterate through the list
		lastBucket = bucket
		bucket = bucket.previous
	}

	// If we reach this, it means the bucketKey is before our last Registered bucketKey
	newBucket := newTimeBucket(bucketKey, nil)
	lastBucket.previous = newBucket
	return newBucket
}

type timeBucket struct {
	bucketKey time.Time
	jobs      map[uuid.UUID]*Job
	previous  *timeBucket
}

func newTimeBucket(bucket time.Time, previous *timeBucket) *timeBucket {
	return &timeBucket{
		jobs:      make(map[uuid.UUID]*Job),
		previous:  previous,
		bucketKey: bucket,
	}
}
