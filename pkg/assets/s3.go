package assets

import (
	"net/url"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/tgracchus/assertuploader/pkg/auerr"
	"github.com/tgracchus/assertuploader/pkg/schedule"
)

const ErrorEmptyAWSCredentials = "ErrorEmptyAWSCredentials"
const ErrorNoAWSCredentials = "ErrorNoAWSCredentials"
const ErrorAlreadyUploaded = "ErrorAlreadyUploaded"

var emptyCredentials = credentials.Credentials{}

func NewS3FileManager(sess *session.Session, scheduler schedule.SimpleScheduler, signedPutExpiration time.Duration) (*S3FileManager, error) {
	svc := s3.New(sess)
	return &S3FileManager{svc: svc, signedPutExpiration: signedPutExpiration, scheduler: scheduler}, nil
}

const metadataPath = "metadata/"
const status = "status"
const uploaded = "uploaded"

type S3FileManager struct {
	svc                 *s3.S3
	signedPutExpiration time.Duration
	scheduler           schedule.SimpleScheduler
}

func (ps *S3FileManager) PutURL(bucket string, assetID uuid.UUID) (*url.URL, error) {
	// Create signed url
	signReq, _ := ps.svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(assetID.String()),
	})

	postURLString, err := signReq.Presign(ps.signedPutExpiration)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}

	postURL, err := url.Parse(postURLString)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}

	tags := url.Values{}
	tags.Set("X-Amz-Expires", postURL.Query().Get("X-Amz-Expires"))
	tags.Set("X-Amz-Date", postURL.Query().Get("X-Amz-Date"))

	// Create signed url timeout mark
	_, err = ps.svc.PutObject(&s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(metadataPath + assetID.String()),
		Tagging: aws.String(tags.Encode()),
	})
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}

	return postURL, nil
}

func (ps *S3FileManager) Uploaded(bucket string, assetID uuid.UUID) error {
	key := assetID.String()
	metadataKey := metadataPath + key
	tags, err := ps.tags(bucket, metadataKey)
	if err != nil {
		return err
	}
	if tag, ok := tags[status]; ok {
		if *tag.Value == uploaded {
			return auerr.FError(ErrorAlreadyUploaded, "Asset %s already uploaded", key)
		}
	}
	err = ps.scheduleJob(bucket, assetID, tags)
	if err != nil {
		return err
	}
	return nil
}

func (ps *S3FileManager) scheduleJob(bucket string, assetID uuid.UUID, tags map[string]*s3.Tag) error {
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
	expirationDate := date.Add((time.Duration(expire) * 2) * time.Second)
	job := schedule.NewFixedDateJob(assetID.String(), ps.newUploadedFunction(bucket, assetID), expirationDate)
	err = ps.scheduler.Schedule(*job)
	if err != nil {
		return err
	}
	return nil
}

func (ps *S3FileManager) newUploadedFunction(bucket string, assetID uuid.UUID) schedule.JobFunction {
	return func() error {
		tags := &s3.Tagging{
			TagSet: []*s3.Tag{
				{Key: aws.String(status), Value: aws.String(uploaded)},
			},
		}
		_, err := ps.svc.PutObjectTagging(
			&s3.PutObjectTaggingInput{
				Bucket:  aws.String(bucket),
				Key:     aws.String(metadataPath + assetID.String()),
				Tagging: tags,
			},
		)
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
}

func (ps *S3FileManager) GetURL(bucket string, assetId uuid.UUID, timeout int) (*url.URL, error) {
	metadataKey := metadataPath + assetId.String()
	tags, err := ps.tags(bucket, metadataKey)
	if err != nil {
		return nil, err
	}
	if tag, ok := tags[status]; ok {
		if *tag.Value != uploaded {
			return nil, auerr.FError(auerr.ErrorNotFound, "Asset %s not marked as uploaded", assetId.String())
		}
	} else {
		return nil, auerr.FError(auerr.ErrorNotFound, "Asset %s not marked as uploaded", assetId.String())

	}

	req, _ := ps.svc.GetObjectRequest(
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(assetId.String()),
		})

	getUrlString, err := req.Presign(time.Duration(timeout) * time.Second)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}

	getUrl, err := url.Parse(getUrlString)
	if err != nil {
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}
	return getUrl, nil

}

func (ps *S3FileManager) tags(bucket string, key string) (map[string]*s3.Tag, error) {
	//Check if it exist
	result, err := ps.svc.GetObjectTagging(
		&s3.GetObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			switch code := awsErr.StatusCode(); code {
			case 404:
				return nil, auerr.FError(auerr.ErrorNotFound, "Asset %s is not found", key)
			default:
				return nil, auerr.CError(auerr.ErrorInternalError, err)
			}
		}
		return nil, auerr.CError(auerr.ErrorInternalError, err)
	}

	tags := make(map[string]*s3.Tag, len(result.TagSet))
	for _, tag := range result.TagSet {
		tags[*tag.Key] = tag

	}
	return tags, nil
}