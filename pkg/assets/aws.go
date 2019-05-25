package assets

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/tgracchus/assetuploader/pkg/auerr"
)

// NewAwsSession creates a new AWS session from the given credentials.
func NewAwsSession(cred *credentials.Credentials, region string) (*session.Session, error) {
	if cred == nil {
		return nil, auerr.SError(auerr.ErrorBadInput, "Credentials are nil")
	}
	if *cred == emptyCredentials {
		return nil, auerr.SError(auerr.ErrorBadInput, "Credentials are empty")
	}
	return session.Must(
		session.NewSession(
			aws.NewConfig().WithCredentials(cred).WithRegion(region),
		),
	), nil
}

// NewS3Client creates a new AWS s3 from the session.
func NewS3Client(sess *session.Session, region string) *s3.S3 {
	return s3.New(sess, aws.NewConfig().WithRegion(region))
}
