package aws_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tgracchus/assertuploader/pkg/aws"
	"log"
	"net/http"
	"strings"
	"testing"
)

const testBucket = "assertuploader"
const testRegion = "eu-west-1"

func TestPoster(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := aws.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := aws.NewS3Manager(session)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PostIt(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("The URL is:", putUrl.String(), " err:", err)
}

func TestUpdateIt(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := aws.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := aws.NewS3Manager(session)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PostIt(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("The URL is:", putUrl.String(), " err:", err)

	req, err := http.NewRequest("PUT", putUrl.String(), strings.NewReader("CONTENT"))
	if err != nil {
		fmt.Println("error creating request", putUrl.String())
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != 200 {
		t.Fatalf("Error put with code %d", response.StatusCode)
	}

	err = poster.UpdateIt(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoubleWrite(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := aws.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := aws.NewS3Manager(session)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PostIt(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("The URL is:", putUrl.String(), " err:", err)

	req, err := http.NewRequest("PUT", putUrl.String(), strings.NewReader("CONTENT"))
	if err != nil {
		fmt.Println("error creating request", putUrl.String())
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("Error put with code %d", response.StatusCode)
	}

	err = poster.UpdateIt(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("PUT", putUrl.String(), strings.NewReader("CONTENT2"))
	if err != nil {
		fmt.Println("error creating request", putUrl.String())
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	response, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("Error put with code %d", response.StatusCode)
	}
	err = poster.UpdateIt(testBucket, assetId)
	if err == nil {
		t.Fatal(errors.New("Second update must fail"))
	}
}

func TestUpdateItFileDoesNotExist(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := aws.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := aws.NewS3Manager(session)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	err = poster.UpdateIt(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPosterUrl(t *testing.T) {
	cred := credentials.NewStaticCredentials("testCredentials", "testSecret", "testKey")
	session, err := aws.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := aws.NewS3Manager(session)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PostIt(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("The URL is:", putUrl.String(), " err:", err)

	expectedHostName := "dmc-asset-uploader-test.s3.us-west-2.amazonaws.com"
	if putUrl.Hostname() != expectedHostName {
		t.Fatalf("Hostname should be %s, not %s", expectedHostName, putUrl.Hostname())
	}

	expectedPath := "/" + assetId.String()
	if putUrl.Path != expectedPath {
		t.Fatalf("Path should be %s, not %s", expectedPath, putUrl.Path)
	}
}

func TestSessionEmptyCredentials(t *testing.T) {
	cred := &credentials.Credentials{}
	_, err := aws.NewAwsSession(testRegion, cred)

	switch code := errors.Cause(err).Error(); code {
	case aws.ErrorEmptyAWSCredentials:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}

func TestSessionNilCredentials(t *testing.T) {
	_, err := aws.NewAwsSession(testRegion, nil)
	switch code := errors.Cause(err).Error(); code {
	case aws.ErrorNoAWSCredentials:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}

func TestPosterEmptyArgs(t *testing.T) {
	cred := credentials.NewStaticCredentials("testCredentials", "testSecret", "testKey")
	session, err := aws.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := aws.NewS3Manager(session)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	_, err = poster.PostIt("", assetId)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch code := awsErr.Code(); code {
			case request.InvalidParameterErrCode:

			default:
				t.Fatalf("We are expecteing an %s not %s", request.InvalidParameterErrCode, code)
			}
		} else {
			t.Fatalf("We are expected an error")
		}
	}
}
