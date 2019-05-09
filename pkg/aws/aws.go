package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"log"
	"net/url"
	"time"
)

func NewAwsSession(region string) *session.Session {
	return session.Must(session.NewSession(
		&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewEnvCredentials(),
		}))
}

func NewPoster(sess *session.Session) (*Poster, error) {
	s3 := s3.New(sess)
	return &Poster{s3}, nil
}

type Poster struct {
	svc *s3.S3
}

func (ps *Poster) PostIt(bucket string, assetId uuid.UUID) (*url.URL, error) {
	req, _ := ps.svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(assetId.String()),
	})
	postUrlString, err := req.Presign(15 * time.Minute)
	if err != nil {
		return nil, err
	}
	postUrl, err := url.Parse(postUrlString)
	if err != nil {
		return nil, err
	}

	log.Println("The URL is:", postUrl.String(), " err:", err)
	return postUrl, nil
}
