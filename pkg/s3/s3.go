package s3

import "github.com/aws/aws-sdk-go"

type PostSigner interface {
	func PostIt()
}


type PostSigner struct