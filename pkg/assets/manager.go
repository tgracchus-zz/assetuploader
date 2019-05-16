package assets

import (
	"context"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/tgracchus/assetuploader/pkg/auerr"
	"github.com/tgracchus/assetuploader/pkg/job"
	"github.com/tgracchus/assetuploader/pkg/schedule"
)

var emptyCredentials = credentials.Credentials{}

const uploadedPath = "uploaded/"
const temporalPath = "temp/"
const status = "status"
const uploaded = "uploaded"

// AssetManager is responsible for the lifecycle of assets.
type AssetManager interface {
	PutURL(ctx context.Context, bucket string, assetID uuid.UUID) (*url.URL, error)
	Uploaded(ctx context.Context, bucket string, assetID uuid.UUID) error
	GetURL(ctx context.Context, bucket string, assetID uuid.UUID, timeout int64) (*url.URL, error)
}

// NewDefaultFileManager creates an AssetManager based on s3 with scheduled execution.
func NewDefaultFileManager(sess *session.Session, region string) AssetManager {
	upsert, query := job.NewMemoryStore(job.MinutesKeys)
	expirationDuration := 30 * time.Second
	scheduler := schedule.NewSimpleScheduler(upsert, query, expirationDuration)
	return News3AssetManager(sess, region, scheduler, expirationDuration)
}

// News3AssetManager creates an AssetManager based on s3 with custom configuration.
func News3AssetManager(sess *session.Session, region string, scheduler schedule.SimpleScheduler, putExpirationTime time.Duration) AssetManager {
	svc := s3.New(sess, aws.NewConfig().WithRegion(region))
	return &s3AssetManager{
		svc:               svc,
		putExpirationTime: putExpirationTime,
		scheduler:         scheduler,
	}
}

type s3AssetManager struct {
	svc               *s3.S3
	putExpirationTime time.Duration
	scheduler         schedule.SimpleScheduler
}

func (ps *s3AssetManager) PutURL(ctx context.Context, bucket string, assetID uuid.UUID) (*url.URL, error) {
	// Create signed url
	signReq, _ := ps.svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(temporalPath + assetID.String()),
	})
	postURLString, err := signReq.Presign(ps.putExpirationTime)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}
	postURL, err := url.Parse(postURLString)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}
	// Create signed mark
	tags := url.Values{}
	tags.Set("X-Amz-Expires", postURL.Query().Get("X-Amz-Expires"))
	tags.Set("X-Amz-Date", postURL.Query().Get("X-Amz-Date"))
	_, err = ps.svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(uploadedPath + assetID.String()),
		Tagging: aws.String(tags.Encode()),
	})
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}
	return postURL, nil
}
func (ps *s3AssetManager) Uploaded(ctx context.Context, bucket string, assetID uuid.UUID) error {
	tags, err := ps.checkIsNotUploaded(ctx, bucket, uploadedPath, assetID)
	if err != nil {
		return err
	}
	expireS := tags["X-Amz-Expires"]
	expire, err := strconv.Atoi(*expireS.Value)
	if err != nil {
		return auerr.CError(auerr.ErrorInternalError, err)
	}
	dateS := tags["X-Amz-Date"]
	date, err := time.Parse("20060102T150405Z0700", *dateS.Value)
	if err != nil {
		return auerr.CError(auerr.ErrorInternalError, err)
	}
	expire = int(math.Round(float64(expire) * 1.10))
	expirationDate := date.Add(time.Duration(expire) * time.Second)
	job := job.NewFixedDateJob(assetID.String(), ps.newUploadedFunction(bucket, assetID), expirationDate)
	return ps.scheduler.Schedule(ctx, *job)
}

func (ps *s3AssetManager) newUploadedFunction(bucket string, assetID uuid.UUID) job.Function {
	return func(ctx context.Context) error {
		// Check if the asset metadata is present and already contains the uploaded tags
		// if its not present, it means not signed Url has been generated
		tags, err := ps.checkIsNotUploaded(ctx, bucket, uploadedPath, assetID)
		if err != nil {
			return err
		}
		// Move the asset to the uploaded folder with proper tags
		updatedTags := url.Values{status: []string{uploaded}}
		for k, v := range tags {
			updatedTags.Add(k, *v.Value)
		}
		_, err = ps.svc.CopyObjectWithContext(
			ctx,
			&s3.CopyObjectInput{
				CopySource:       aws.String(bucket + "/" + temporalPath + assetID.String()),
				Bucket:           aws.String(bucket),
				Key:              aws.String(uploadedPath + assetID.String()),
				Tagging:          aws.String(updatedTags.Encode()),
				TaggingDirective: aws.String(s3.TaggingDirectiveReplace),
			},
		)
		return ps.handleAwsError(err, assetID)
	}
}

func (ps *s3AssetManager) GetURL(ctx context.Context, bucket string, assetID uuid.UUID, timeout int64) (*url.URL, error) {
	_, err := ps.checkIsUploaded(ctx, bucket, uploadedPath, assetID)
	if err != nil {
		return nil, err
	}
	req, _ := ps.svc.GetObjectRequest(
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(uploadedPath + assetID.String()),
		})

	getURLString, err := req.Presign(time.Duration(timeout) * time.Second)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}

	getURL, err := url.Parse(getURLString)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}
	return getURL, nil

}
func (ps *s3AssetManager) handleAwsError(err error, assetID uuid.UUID) error {
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			switch code := awsErr.StatusCode(); code {
			case 404:
				return auerr.FError(auerr.ErrorNotFound, "Asset %s is not found", assetID.String())
			default:
				return auerr.CError(auerr.ErrorInternalError, err)
			}
		}
		return auerr.CError(auerr.ErrorInternalError, err)
	}
	return nil
}
func (ps *s3AssetManager) checkIsUploaded(ctx context.Context, bucket string, path string, assetID uuid.UUID) (map[string]*s3.Tag, error) {
	tags, err := ps.tags(ctx, bucket, path, assetID)
	if err != nil {
		return nil, err
	}
	if tag, ok := tags[status]; ok {
		if *tag.Value == uploaded {
			return tags, nil
		}
	}
	return nil, auerr.FError(auerr.ErrorConflict, "Asset %s already uploaded", assetID.String())
}

func (ps *s3AssetManager) checkIsNotUploaded(ctx context.Context, bucket string, path string, assetID uuid.UUID) (map[string]*s3.Tag, error) {
	tags, err := ps.tags(ctx, bucket, path, assetID)
	if err != nil {
		return nil, err
	}
	if tag, ok := tags[status]; ok {
		if *tag.Value == uploaded {
			return nil, auerr.FError(auerr.ErrorConflict, "Asset %s already uploaded", assetID.String())
		}
	}
	return tags, nil
}
func (ps *s3AssetManager) tags(ctx context.Context, bucket string, path string, assetID uuid.UUID) (map[string]*s3.Tag, error) {
	//Check if it exist
	result, err := ps.svc.GetObjectTaggingWithContext(
		ctx,
		&s3.GetObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(path + assetID.String()),
		},
	)
	err = ps.handleAwsError(err, assetID)
	if err != nil {
		return nil, err
	}
	tags := make(map[string]*s3.Tag, len(result.TagSet))
	for _, tag := range result.TagSet {
		tags[*tag.Key] = tag

	}
	return tags, nil
}
