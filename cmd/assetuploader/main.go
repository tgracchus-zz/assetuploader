package main

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/labstack/echo"
	"github.com/tgracchus/assertuploader/pkg/assets"
	"github.com/tgracchus/assertuploader/pkg/endpoints"
)

func main() {
	pflag.String("region", "us-west-2", "aws region")
	pflag.String("bucket", "dmc-asset-uploader-test", "aws bucket")
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.BindPFlags(pflag.CommandLine)
	pflag.Parse()
	region := viper.GetString("region")
	bucket := viper.GetString("bucket")
	// Setted as env variables only, so we dont see credentials in cmd history
	awsKey := viper.GetString("AWS_ACCESS_KEY_ID")
	if awsKey == "" {
		panic("AWS_ACCESS_KEY_ID should be present in env vars")
	}
	awsSecret := viper.GetString("AWS_SECRET_ACCESS_KEY")
	if awsSecret == "" {
		panic("AWS_SECRET_ACCESS_KEY should be present in env vars")
	}
	e := echo.New()
	e.HTTPErrorHandler = endpoints.AssetUploaderHTTPErrorHandler
	credentials := credentials.NewStaticCredentials(awsKey, awsSecret, "")
	session, err := assets.NewAwsSession(credentials)
	if err != nil {
		panic(err)
	}
	manager := assets.NewDefaultFileManager(session, region)
	endpoints.RegisterAssetsEndpoints(e, manager, bucket)
	e.Logger.Fatal(e.Start(":8080"))
}
