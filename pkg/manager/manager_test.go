package manager_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tgracchus/assertuploader/pkg/manager"
	"github.com/tgracchus/assertuploader/pkg/scheduler"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

const testBucket = "assertuploader"
const testRegion = "eu-west-1"

func TestPoster(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := manager.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := manager.NewS3Manager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PutURL(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("The URL is:", putUrl.String(), " err:", err)
}

func TestUpdateIt(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := manager.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := manager.NewS3Manager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PutURL(testBucket, assetId)
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

	err = poster.Uploaded(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoubleWrite(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := manager.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}

	signedPutExpiration := 2 * time.Second
	poster, err := manager.NewS3Manager(session,
		scheduler.NewSimpleScheduler(
			scheduler.NewMemoryJobStore(func(date time.Time) time.Time {
				return date.Truncate(time.Second)
			}),
			signedPutExpiration/2,
		),
		signedPutExpiration,
	)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PutURL(testBucket, assetId)
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

	err = poster.Uploaded(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	getUrl, err := poster.GetURL(testBucket, assetId, 15)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", getUrl.String(), nil)
	if err != nil {
		fmt.Println("error creating request", putUrl.String())
		return
	}
	response, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("Error put with code %d", response.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	body := string(bodyBytes)

	if body != "CONTENT" {
		t.Fatalf("Body should be CONTENT, not %s", body)
	}

	log.Println("The URL is:", getUrl.String(), " err:", err)

}

func TestUpdateItFileDoesNotExist(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := manager.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := manager.NewS3Manager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	err = poster.Uploaded(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPosterUrl(t *testing.T) {
	cred := credentials.NewStaticCredentials("testCredentials", "testSecret", "testKey")
	session, err := manager.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := manager.NewS3Manager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := poster.PutURL(testBucket, assetId)
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
	_, err := manager.NewAwsSession(testRegion, cred)

	switch code := errors.Cause(err).Error(); code {
	case manager.ErrorEmptyAWSCredentials:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}

func TestSessionNilCredentials(t *testing.T) {
	_, err := manager.NewAwsSession(testRegion, nil)
	switch code := errors.Cause(err).Error(); code {
	case manager.ErrorNoAWSCredentials:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}

func TestPosterEmptyArgs(t *testing.T) {
	cred := credentials.NewStaticCredentials("testCredentials", "testSecret", "testKey")
	session, err := manager.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := manager.NewS3Manager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	_, err = poster.PutURL("", assetId)
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
