package assets_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tgracchus/assertuploader/pkg/assets"
)

const testBucket = "dmc-asset-uploader-test"
const testRegion = "​ ​us-west-2"

func TestPutUrl(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := assets.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := assets.NewS3FileManager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	putUrl, err := manager.PutURL(testBucket, assetId)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("The URL is:", putUrl.String(), " err:", err)
}

func TestUpdateIt(t *testing.T) {
	cred := credentials.NewEnvCredentials()
	session, err := assets.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := assets.NewS3FileManager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
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
	session, err := assets.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}

	signedPutExpiration := 2 * time.Second
	poster, err := assets.NewS3FileManager(session,
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
	session, err := assets.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := assets.NewS3FileManager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
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
	session, err := assets.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := assets.NewS3FileManager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
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
	_, err := assets.NewAwsSession(testRegion, cred)

	switch code := errors.Cause(err).Error(); code {
	case assets.ErrorEmptyAWSCredentials:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}

func TestSessionNilCredentials(t *testing.T) {
	_, err := assets.NewAwsSession(testRegion, nil)
	switch code := errors.Cause(err).Error(); code {
	case assets.ErrorNoAWSCredentials:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}

func TestPosterEmptyArgs(t *testing.T) {
	cred := credentials.NewStaticCredentials("testCredentials", "testSecret", "testKey")
	session, err := assets.NewAwsSession(testRegion, cred)
	if err != nil {
		t.Fatal(err)
	}
	poster, err := assets.NewS3FileManager(session, scheduler.NewImmediateScheduler(), 2*time.Minute)
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
