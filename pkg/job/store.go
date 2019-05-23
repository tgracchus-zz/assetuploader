package job

import (
	"context"
	"time"

	"github.com/tgracchus/assetuploader/pkg/auerr"
)

// NewMemoryStore instantiates a new store in memory storage.
func NewMemoryStore(bucketKeyFunc BucketKeyFunc) (chan Job, chan StoreQuery) {
	upSert := make(chan Job, 1000)
	queries := make(chan StoreQuery, 1000)
	jobs := newTimeBuckets(bucketKeyFunc)
	go func() {
		defer close(upSert)
		defer close(queries)
		for {
			select {
			case job, ok := <-upSert:
				if !ok {
					upSert = nil
				}
				jobs.upsert(job)
			case query, ok := <-queries:
				if !ok {
					queries = nil
				}
				jobs := jobs.findBucketsBefore(query)
				select {
				case <-query.ctx.Done():
					//Do nothing channel context is aborted
				default:
					query.response <- jobs
				}

			}
			if upSert == nil || queries == nil {
				panic("Upsert or queries closes")
			}
		}
	}()
	return upSert, queries
}

// UpSert sends a job to the upset channel of a store.
func UpSert(ctx context.Context, upSert chan Job, job Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		upSert <- job
	}
	return nil
}

// GetBeforeCriteria sets the criteria to add a job to search results by the GetBefore method.
type GetBeforeCriteria func(jobs Job) bool

// GetBefore ask for jobs whith status and with execution date before than the given one.
func GetBefore(ctx context.Context, queries chan StoreQuery, date time.Time, criteria GetBeforeCriteria) ([]Job, error) {
	query := StoreQuery{
		ctx:      ctx,
		date:     date,
		response: make(chan []Job),
		criteria: criteria}

	defer close(query.response)
	queries <- query
	select {
	case response, ok := <-query.response:
		if !ok {
			return nil, auerr.SError(auerr.ErrorInternalError, "Can not get jobs")
		}
		return response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// StoreQuery struct to query for jobs with status and with executionDate before date.
type StoreQuery struct {
	ctx      context.Context
	date     time.Time
	response chan []Job
	criteria GetBeforeCriteria
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

func (j *jobs) findBucketsBefore(query StoreQuery) []Job {
	bucketKey := j.bucketKeyFunc(query.date)
	bucket := j.headBucket
	jobs := make([]Job, 0, 0)
	for bucket != nil {
		if bucket.bucketKey <= bucketKey {
			for _, job := range bucket.Jobs {
				if ok := query.criteria(job); ok {
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
