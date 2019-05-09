package aws_test

import (
	"github.com/google/uuid"
	"github.com/tgracchus/assertuploader/pkg/aws"
	"os"
	"testing"
)

func TestPoster(t *testing.T) {

	os.Setenv("AWS_ACCESS_KEY_ID", "1asdasd")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "1asdasd")

	session := aws.NewAwsSession("us-west-2")
	poster, err := aws.NewPoster(session)
	if err != nil {
		t.Fatal(err)
	}
	assetId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	post, err := poster.PostIt("dmc-asset-uploader-test", assetId)
	if err != nil {
		t.Fatal(err)
	}

	expectedHostName := "dmc-asset-uploader-test.s3.us-west-2.amazonaws.com"
	if post.Hostname() != expectedHostName {
		t.Fatalf("Hostname should be %s, not %s", expectedHostName, post.Hostname())
	}

	expectedPath := "/" + assetId.String()
	if post.Path != expectedPath {
		t.Fatalf("Path should be %s, not %s", expectedPath, post.Path)
	}
}
