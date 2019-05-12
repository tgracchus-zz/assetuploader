package manager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/tgracchus/assertuploader/pkg/auerrors"
	"github.com/tgracchus/assertuploader/pkg/scheduler"
	"net/url"
	"strconv"
	"time"
)

const ErrorEmptyAWSCredentials = "ErrorEmptyAWSCredentials"
const ErrorNoAWSCredentials = "ErrorNoAWSCredentials"
const ErrorAlreadyUploaded = "ErrorAlreadyUploaded"

var emptyCredentials = credentials.Credentials{}

func NewAwsSession(region string, cred *credentials.Credentials) (*session.Session, error) {
	if cred == nil {
		return nil, auerrors.SError(ErrorNoAWSCredentials, "Credentials are nil")
	}
	if *cred == emptyCredentials {
		return nil, auerrors.SError(ErrorEmptyAWSCredentials, "Credentials are empty")
	}
	return session.Must(session.NewSession(
		&aws.Config{
			Region:      aws.String(region),
			Credentials: cred,
		})), nil
}

func NewS3Manager(sess *session.Session, scheduler scheduler.SimpleScheduler, signedPutExpiration time.Duration) (*S3Manager, error) {
	svc := s3.New(sess)
	return &S3Manager{svc: svc, signedPutExpiration: signedPutExpiration, scheduler: scheduler}, nil
}

type S3Manager struct {
	svc                 *s3.S3
	signedPutExpiration time.Duration
	scheduler           scheduler.SimpleScheduler
}

/*
func (ps *S3Manager) PostItA(bucket string, assetId uuid.UUID) (*url.URL, error) {
	// Create signed url
	signer := v4.NewSigner(credentials.NewEnvCredentials())
	req := httptest.NewRequest("PUT", "https://assertuploader.s3.eu-west-1.amazonaws.com/"+assetId.String(), strings.NewReader("CONTENT"))
	header, err := signer.Presign(req, nil, "s3", "eu-west-1", 15*time.Minute, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "Presign Parse URL Error")
	}
	req.Header = header

	return req.URL, nil
}
*/

func (ps *S3Manager) PutURL(bucket string, assetId uuid.UUID) (*url.URL, error) {
	// Create signed url
	signReq, _ := ps.svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(assetId.String()),
	})

	postUrlString, err := signReq.Presign(ps.signedPutExpiration)
	if err != nil {
		return nil, auerrors.CError(auerrors.ErrorInternalError, err)
	}

	postUrl, err := url.Parse(postUrlString)
	if err != nil {
		return nil, auerrors.CError(auerrors.ErrorInternalError, err)
	}

	tags := url.Values{}
	tags.Set("X-Amz-Expires", postUrl.Query().Get("X-Amz-Expires"))
	tags.Set("X-Amz-Date", postUrl.Query().Get("X-Amz-Date"))

	// Create signed url timeout mark
	_, err = ps.svc.PutObject(&s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String("metadata/" + assetId.String()),
		Tagging: aws.String(tags.Encode()),
	})
	if err != nil {
		return nil, auerrors.CError(auerrors.ErrorInternalError, err)
	}

	return postUrl, nil
}

func (ps *S3Manager) Uploaded(bucket string, assetId uuid.UUID) error {
	key := assetId.String()
	metadataKey := "/metadata/" + key
	tags, err := ps.tags(bucket, metadataKey)
	if err != nil {
		return err
	}
	if tag, ok := tags["status"]; ok {
		if *tag.Value == "uploaded" {
			return auerrors.FError(ErrorAlreadyUploaded, "Asset %s already uploaded", key)
		}
	}
	err = ps.scheduleJob(bucket, assetId, tags)
	if err != nil {
		return err
	}
	return nil
}

func (ps *S3Manager) scheduleJob(bucket string, assetId uuid.UUID, tags map[string]*s3.Tag) error {
	expireS := tags["X-Amz-Expires"]
	expire, err := strconv.Atoi(*expireS.Value)
	if err != nil {
		return auerrors.CError(auerrors.ErrorInternalError, err)
	}
	dateS := tags["X-Amz-Date"]
	date, err := time.Parse("20060102T150405Z0700", *dateS.Value)
	if err != nil {
		return auerrors.CError(auerrors.ErrorInternalError, err)
	}
	expirationDate := date.Add((time.Duration(expire) * 2) * time.Second)
	job := scheduler.NewFixedDateJob(assetId, ps.newUploadedFunction(bucket, assetId), expirationDate)
	err = ps.scheduler.Schedule(job)
	if err != nil {
		return err
	}
	return nil
}

func (ps *S3Manager) newUploadedFunction(bucket string, assetId uuid.UUID) scheduler.JobFunction {
	return func() error {
		tags := &s3.Tagging{
			TagSet: []*s3.Tag{
				{Key: aws.String("Status"), Value: aws.String("uploaded")},
			},
		}
		_, err := ps.svc.PutObjectTagging(
			&s3.PutObjectTaggingInput{
				Bucket:  aws.String(bucket),
				Key:     aws.String("metadata/" + assetId.String()),
				Tagging: tags,
			},
		)
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok {
				switch code := awsErr.StatusCode(); code {
				case 404:
					return auerrors.FError(auerrors.ErrorNotFound, "Asset %s is not found", assetId.String())
				default:
					return auerrors.CError(auerrors.ErrorInternalError, err)
				}
			}
			return auerrors.CError(auerrors.ErrorInternalError, err)
		}

		return nil
	}
}

func (ps *S3Manager) GetURL(bucket string, assetId uuid.UUID, timeout int) (*url.URL, error) {
	metadataKey := "/metadata/" + assetId.String()
	tags, err := ps.tags(bucket, metadataKey)
	if err != nil {
		return nil, err
	}
	if tag, ok := tags["Status"]; ok {
		if *tag.Value != "uploaded" {
			return nil, auerrors.FError(auerrors.ErrorNotFound, "Asset %s not marked as uploaded", assetId.String())
		}
	} else {
		return nil, auerrors.FError(auerrors.ErrorNotFound, "Asset %s not marked as uploaded", assetId.String())

	}

	req, _ := ps.svc.GetObjectRequest(
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(assetId.String()),
		})

	getUrlString, err := req.Presign(time.Duration(timeout) * time.Second)
	if err != nil {
		return nil, auerrors.CError(auerrors.ErrorInternalError, err)
	}

	getUrl, err := url.Parse(getUrlString)
	if err != nil {
		return nil, auerrors.CError(auerrors.ErrorInternalError, err)
	}
	return getUrl, nil

}

func (ps *S3Manager) tags(bucket string, key string) (map[string]*s3.Tag, error) {
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
				return nil, auerrors.FError(auerrors.ErrorNotFound, "Asset %s is not found", key)
			default:
				return nil, auerrors.CError(auerrors.ErrorInternalError, err)
			}
		}
		return nil, auerrors.CError(auerrors.ErrorInternalError, err)
	}

	tags := make(map[string]*s3.Tag, len(result.TagSet))
	for _, tag := range result.TagSet {
		tags[*tag.Key] = tag

	}
	return tags, nil
}
