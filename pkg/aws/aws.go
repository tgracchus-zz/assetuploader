package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tgracchus/assertuploader/pkg/auerrors"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
)

const ErrorEmptyAWSCredentials = "ErrorEmptyAWSCredentials"
const ErrorNoAWSCredentials = "ErrorNoAWSCredentials"

var emptyCredentials = credentials.Credentials{}

func NewAwsSession(region string, cred *credentials.Credentials) (*session.Session, error) {
	if cred == nil {
		return nil, auerrors.New(ErrorNoAWSCredentials, "Credentials are nil")
	}
	if *cred == emptyCredentials {
		return nil, auerrors.New(ErrorEmptyAWSCredentials, "Credentials are empty")
	}
	return session.Must(session.NewSession(
		&aws.Config{
			Region:      aws.String(region),
			Credentials: cred,
		})), nil
}

func NewS3Manager(sess *session.Session) (*S3Manager, error) {
	scv := s3.New(sess)
	return &S3Manager{scv}, nil
}

type S3Manager struct {
	svc *s3.S3
}

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

const ErrorPostURLAWS = "ErrorPostURLAWS"

func (ps *S3Manager) PutURL(bucket string, assetId uuid.UUID) (*url.URL, error) {
	// Create signed url
	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("pending/" + assetId.String()),
	}
	req, _ := ps.svc.PutObjectRequest(putObjectInput)
	postUrlString, err := req.Presign(15 * time.Minute)
	if err != nil {
		return nil, auerrors.NewWithError(ErrorPostURLAWS, err)
	}

	postUrl, err := url.Parse(postUrlString)
	if err != nil {
		return nil, auerrors.NewWithError(ErrorPostURLAWS, err)
	}
	return postUrl, nil
}

const ErrorUpdateItAWS = "ErrorUpdateItAWS"
const ErrorAlreadyUploaded = "ErrorAlreadyUploaded"

func (ps *S3Manager) UpdateIt(bucket string, assetId uuid.UUID) error {
	key := assetId.String()
	uploadedKey := "/uploaded/" + key
	pendingKey := "/pending/" + key

	uploadedHead, err := ps.head(bucket, uploadedKey)
	if err != nil {
		code := errors.Cause(err).Error()
		if code != ErrorNotFound {
			return err
		}
	}

	if uploadedHead != nil {
		return auerrors.New(ErrorAlreadyUploaded, fmt.Sprintf("Asset %s is already uploaded", assetId))
	}

	pendingHead, err := ps.head(bucket, pendingKey)
	if err != nil {
		return err
	}
	_, err = ps.svc.CopyObject(
		&s3.CopyObjectInput{
			Bucket:            aws.String(bucket),
			Key:               aws.String(uploadedKey),
			CopySource:        aws.String(bucket + pendingKey),
			CopySourceIfMatch: pendingHead.ETag,
		},
	)
	if err != nil {
		return auerrors.NewWithError(ErrorUpdateItAWS, err)
	}

	return nil
}

const ErrorGetURLAWS = "ErrorGetURLAWS"

func (ps *S3Manager) GetURL(bucket string, assetId uuid.UUID, timeout int) (*url.URL, error) {
	key := assetId.String()
	uploadedKey := "/uploaded/" + key

	req, _ := ps.svc.GetObjectRequest(
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(uploadedKey),
		})

	getUrlString, err := req.Presign(time.Duration(timeout) * time.Second)
	if err != nil {
		return nil, auerrors.NewWithError(ErrorGetURLAWS, err)
	}

	getUrl, err := url.Parse(getUrlString)
	if err != nil {
		return nil, auerrors.NewWithError(ErrorGetURLAWS, err)
	}
	return getUrl, nil

}

const ErrorHeadAWS = "ErrorHeadAWS"
const ErrorNotFound = "ErrorNotFound"

func (ps *S3Manager) head(bucket string, key string) (*s3.HeadObjectOutput, error) {
	//Check if it exist
	head, err := ps.svc.HeadObject(
		&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok {
			switch code := awsErr.StatusCode(); code {
			case 404:
				return nil, auerrors.New(ErrorNotFound, fmt.Sprintf("Asset %s is not found", key))
			default:
				return nil, auerrors.NewWithError(ErrorHeadAWS, err)
			}
		}
		return nil, auerrors.NewWithError(ErrorHeadAWS, err)
	}

	return head, nil
}
