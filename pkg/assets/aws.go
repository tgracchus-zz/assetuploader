package assets

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/tgracchus/assetuploader/pkg/auerr"
)

// NewAwsSession creates a new AWS session from the given credentials.
func NewAwsSession(cred *credentials.Credentials) (*session.Session, error) {
	if cred == nil {
		return nil, auerr.SError(auerr.ErrorBadInput, "Credentials are nil")
	}
	if *cred == emptyCredentials {
		return nil, auerr.SError(auerr.ErrorBadInput, "Credentials are empty")
	}
	return session.Must(session.NewSession(
		&aws.Config{
			Credentials: cred,
		})), nil
}
