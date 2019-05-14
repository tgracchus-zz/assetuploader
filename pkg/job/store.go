package job

import (
	"time"
)

// Store is a time aware job store, it´s able to retrieve all the jobs with the given status
// and meant to be executed a given date.
type Store interface {
	UpSert(job Job) error
	GetBefore(date time.Time, statuses []Status) ([]Job, error)
}

// NewMemoryStore it´s an in memory Store
func NewMemoryStore(bucketKeyFunc BucketKeyFunc) Store {
	upSert := make(chan Job, 1000)
	query := make(chan getBefore, 1000)
	out := make(chan []Job, 1000)
	store := &memoryJobStore{upSert: upSert, query: query, out: out}
	memoryJobStoreMonitor(bucketKeyFunc, upSert, query, out)
	return store
}

type memoryJobStore struct {
	upSert chan<- Job
	query  chan<- getBefore
	out    <-chan []Job
}

func (st *memoryJobStore) UpSert(job Job) error {
	st.upSert <- job
	return nil
}

func (st *memoryJobStore) GetBefore(date time.Time, statuses []Status) ([]Job, error) {
	statusesMap := make(map[Status]bool)
	for _, status := range statuses {
		statusesMap[status] = true
	}
	st.query <- getBefore{date: date, status: statusesMap}
	return <-st.out, nil
}

type getBefore struct {
	date   time.Time
	status map[Status]bool
}

func memoryJobStoreMonitor(bucketKeyFunc BucketKeyFunc, upSert chan Job, query chan getBefore, out chan []Job) {
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
				out <- jobs.getBefore(get)
			}
			if upSert == nil && query == nil {
				break
			}
		}
	}()
}

func newTimeBuckets(bucketKeyFunc BucketKeyFunc) *jobs {
	now := bucketKeyFunc(time.Now())
	bucket := newTimeBucket(now, nil)
	buckets := make(map[string]timeBucket)
	buckets[now] = bucket
	return &jobs{buckets, &bucket, bucketKeyFunc}
}

type jobs struct {
	Buckets       map[string]timeBucket `json:"buckets"`
	headBucket    *timeBucket
	bucketKeyFunc BucketKeyFunc
}

func (j *jobs) upsert(job Job) {
	now := j.bucketKeyFunc(job.ExecutionDate)
	bucket := j.findOrCreateBucket(now)
	bucket.Jobs[job.ID] = job
}

func (j *jobs) getBefore(before getBefore) []Job {
	return j.findBucketsBefore(before)
}

func (j *jobs) findBucketsBefore(before getBefore) []Job {
	buketKey := j.bucketKeyFunc(before.date)
	bucket := j.headBucket
	jobs := make([]Job, 0, 0)
	for bucket != nil {
		if bucket.bucketKey <= buketKey {
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

func (j *jobs) findOrCreateBucket(bucketKey string) *timeBucket {
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
	bucketKey string         `json:"-"`
	Jobs      map[string]Job `json:"jobs"`
	previous  *timeBucket    `json:"-"`
}

func newTimeBucket(bucket string, previous *timeBucket) timeBucket {
	return timeBucket{
		Jobs:      make(map[string]Job),
		previous:  previous,
		bucketKey: bucket,
	}
}

// BucketKeyFunc used by the in memory Store to adjust the granurality of
// the time buckets where the jobs are store.
type BucketKeyFunc func(date time.Time) string

// MillisKeys BucketKeyFunc with Milliseconds granurality.
func MillisKeys(date time.Time) string {
	return date.Truncate(time.Millisecond).UTC().Format(time.RFC3339)
}

// SecondsKeys BucketKeyFunc with Seconds granurality.
func SecondsKeys(date time.Time) string {
	return date.Truncate(time.Second).UTC().Format(time.RFC3339)
}

// MinutesKeys BucketKeyFunc with Minutes granurality.
func MinutesKeys(date time.Time) string {
	return date.Truncate(time.Minute).UTC().Format(time.RFC3339)
}
