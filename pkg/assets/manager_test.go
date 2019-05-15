package assets_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tgracchus/assetuploader/pkg/assets"
	"github.com/tgracchus/assetuploader/pkg/auerr"
	"github.com/tgracchus/assetuploader/pkg/job"
	"github.com/tgracchus/assetuploader/pkg/schedule"
)

// TestS3AssetManager requires to set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY env variables.
func TestS3AssetManager(t *testing.T) {
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("AWS_BUCKET")
	credentials := credentials.NewEnvCredentials()
	session, err := assets.NewAwsSession(credentials, region)
	if err != nil {
		t.Fatal(err)
	}
	manager := assets.News3AssetManager(session, region, schedule.NewImmediateScheduler(), 2*time.Minute)
	t.Run("TestUpdateIt", newTestUpdateIt(manager, bucket))
	t.Run("TestOverwrite", newTestOverwrite(session, bucket, region))
	t.Run("TestUpdateItFileDoesNotExist", newTestUpdateItFileDoesNotExist(manager, bucket))
	t.Run("TestPutUrl", newTestPutUrl(manager, bucket, region))

}

func newTestUpdateIt(manager assets.AssetManager, bucket string) func(t *testing.T) {
	return func(t *testing.T) {
		assetId, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}
		putUrl, err := manager.PutURL(bucket, assetId)
		if err != nil {
			t.Fatal(err)
		}
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
		err = manager.Uploaded(bucket, assetId)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Second)
		getUrl, err := manager.GetURL(bucket, assetId, 15)
		if err != nil {
			t.Fatal(err)
		}
		req, err = http.NewRequest("GET", getUrl.String(), nil)
		if err != nil {
			fmt.Println("error creating request", getUrl.String())
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
	}
}

func newTestOverwrite(session *session.Session, bucket string, region string) func(t *testing.T) {
	return func(t *testing.T) {
		signedPutExpiration := 2 * time.Second
		mamager := assets.News3AssetManager(session, region,
			schedule.NewSimpleScheduler(
				job.NewMemoryStore(job.SecondsKeys),
				signedPutExpiration/2,
			),
			signedPutExpiration,
		)
		assetId, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}
		putURL, err := mamager.PutURL(bucket, assetId)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", putURL.String(), strings.NewReader("CONTENT"))
		if err != nil {
			fmt.Println("error creating request", putURL.String())
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
		err = mamager.Uploaded(bucket, assetId)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Second)
		getUrl, err := mamager.GetURL(bucket, assetId, 15)
		if err != nil {
			t.Fatal(err)
		}
		req, err = http.NewRequest("GET", getUrl.String(), nil)
		if err != nil {
			fmt.Println("error creating request", getUrl.String())
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
	}
}
func newTestUpdateItFileDoesNotExist(manager assets.AssetManager, bucket string) func(t *testing.T) {
	return func(t *testing.T) {
		assetId, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}
		err = manager.Uploaded(bucket, assetId)
		if err == nil {
			t.Fatal("We expected and error")
		}
	}
}

func newTestPutUrl(manager assets.AssetManager, bucket string, region string) func(t *testing.T) {
	return func(t *testing.T) {
		assetId, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}
		putURL, err := manager.PutURL(bucket, assetId)
		if err != nil {
			t.Fatal(err)
		}
		expectedHostName := bucket + ".s3." + region + ".amazonaws.com"
		if putURL.Hostname() != expectedHostName {
			t.Fatalf("Hostname should be %s, not %s", expectedHostName, putURL.Hostname())
		}
		expectedPath := "/temp/" + assetId.String()
		if putURL.Path != expectedPath {
			t.Fatalf("Path should be %s, not %s", expectedPath, putURL.Path)
		}
	}
}

func TestSessionEmptyCredentials(t *testing.T) {
	cred := &credentials.Credentials{}
	region := os.Getenv("TEST_REGION")
	_, err := assets.NewAwsSession(cred, region)

	switch code := errors.Cause(err).Error(); code {
	case auerr.ErrorBadInput:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}

func TestSessionNilCredentials(t *testing.T) {
	region := os.Getenv("TEST_REGION")
	_, err := assets.NewAwsSession(nil, region)
	switch code := errors.Cause(err).Error(); code {
	case auerr.ErrorBadInput:
	default:
		t.Fatalf("We are expected an auerrors.AUError")
	}
}
