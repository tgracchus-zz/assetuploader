package job

import (
	"time"
)

// NewStore instantiates a new store with the given store function.
func NewStore(store Store) (chan Job, chan StoreQuery, chan []Job) {
	upSert := make(chan Job, 1000)
	query := make(chan StoreQuery, 1000)
	out := make(chan []Job, 1000)
	store(upSert, query, out)
	return upSert, query, out
}

// UpSert sends a job to the upset channel of a store.
func UpSert(upSert chan Job, job Job) error {
	upSert <- job
	return nil
}

// GetBefore ask for jobs whith status and with execution date before than the given one.
func GetBefore(query chan StoreQuery, out chan []Job, date time.Time, statuses []Status) ([]Job, error) {
	statusesMap := make(map[Status]bool)
	for _, status := range statuses {
		statusesMap[status] = true
	}
	query <- StoreQuery{date: date, status: statusesMap}
	return <-out, nil
}

// StoreQuery struct to query for jobs with status and with executionDate before date.
type StoreQuery struct {
	date   time.Time
	status map[Status]bool
}

// Store is a function for storing and look for jobs
type Store func(upSert chan Job, query chan StoreQuery, storeQueryResult chan []Job)

// NewMemoryStore creates an in memory Store using time buckets to classify jobs.
func NewMemoryStore(bucketKeyFunc BucketKeyFunc) Store {
	return func(upSert chan Job, query chan StoreQuery, storeQueryResult chan []Job) {
		jobs := newTimeBuckets(bucketKeyFunc)
		go func() {
			for {
				select {
				case job, ok := <-upSert:
					if !ok {
						upSert = nil
					}
					jobs.upsert(job)
				case get, ok := <-query:
					if !ok {
						query = nil
					}
					storeQueryResult <- jobs.getBefore(get)
				}
				if upSert == nil && query == nil {
					break
				}
			}
		}()
	}
}

func newTimeBuckets(bucketKeyFunc BucketKeyFunc) *jobs {
	now := bucketKeyFunc(time.Now())
	bucket := newTimeBucket(now, nil)
	buckets := make(map[int64]timeBucket)
	buckets[now] = bucket
	return &jobs{buckets, &bucket, bucketKeyFunc}
}

type jobs struct {
	Buckets       map[int64]timeBucket `json:"buckets"`
	headBucket    *timeBucket
	bucketKeyFunc BucketKeyFunc
}

func (j *jobs) upsert(job Job) {
	now := j.bucketKeyFunc(job.ExecutionDate)
	bucket := j.findOrCreateBucket(now)
	bucket.Jobs[job.ID] = job
}

func (j *jobs) getBefore(before StoreQuery) []Job {
	return j.findBucketsBefore(before)
}

func (j *jobs) findBucketsBefore(before StoreQuery) []Job {
	bucketKey := j.bucketKeyFunc(before.date)
	bucket := j.headBucket
	jobs := make([]Job, 0, 0)
	for bucket != nil {
		if bucket.bucketKey <= bucketKey {
			for _, job := range bucket.Jobs {
				if ok := before.status[job.Status]; ok {
					jobs = append(jobs, job)
				}
			}
		}
		bucket = bucket.previous
	}
	return jobs
}

func (j *jobs) findOrCreateBucket(bucketKey int64) *timeBucket {
	bucket := j.headBucket
	var lastBucket *timeBucket
	for bucket != nil {
		//Bucket has the same key than current bucket, just return it
		if bucket.bucketKey == bucketKey {
			return bucket
		}
		//Bucket is newer than the current bucket, create it
		if bucketKey > bucket.bucketKey {
			newBucket := newTimeBucket(bucketKey, bucket)
			newBucket.previous = bucket
			//if previous bucket is nil it means we need to change the headBucket
			if lastBucket == nil {
				j.headBucket = &newBucket
			}
			return &newBucket
		}

		// Just iterate through the list
		lastBucket = bucket
		bucket = bucket.previous
	}

	// If we reach this, it means the bucketKey is before our last Registered bucketKey
	newBucket := newTimeBucket(bucketKey, nil)
	lastBucket.previous = &newBucket
	return &newBucket
}

type timeBucket struct {
	bucketKey int64          `json:"-"`
	Jobs      map[string]Job `json:"jobs"`
	previous  *timeBucket    `json:"-"`
}

func newTimeBucket(bucket int64, previous *timeBucket) timeBucket {
	return timeBucket{
		Jobs:      make(map[string]Job),
		previous:  previous,
		bucketKey: bucket,
	}
}

// BucketKeyFunc used by the in memory Store to adjust the granurality of
// the time buckets where the jobs are store.
type BucketKeyFunc func(date time.Time) int64

// MillisKeys BucketKeyFunc with Milliseconds granurality.
func MillisKeys(date time.Time) int64 {
	return date.Truncate(time.Millisecond).UTC().Unix()
}

// SecondsKeys BucketKeyFunc with Seconds granurality.
func SecondsKeys(date time.Time) int64 {
	return date.Truncate(time.Second).UTC().Unix()
}

// MinutesKeys BucketKeyFunc with Minutes granurality.
func MinutesKeys(date time.Time) int64 {
	return date.Truncate(time.Minute).UTC().Unix()
}
