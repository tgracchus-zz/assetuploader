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
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	pflag.String("region", "us-west-2", "aws region")
	pflag.String("bucket", "dmc-asset-uploader-test", "aws bucket")
	pflag.Parse()

	region := viper.GetString("region")
	bucket := viper.GetString("bucket")
	// Setted as env variables only, so we dont see credentials in cmd history
	awsKey := viper.GetString("AWS_ACCESS_KEY_ID")
	awsSecret := viper.GetString("AWS_SECRET_ACCESS_KEY")

	e := echo.New()
	e.HTTPErrorHandler = AssetUploaderHTTPErrorHandler()
	credentials := credentials.NewStaticCredentials(awsKey, awsSecret, "")
	session, err := assets.NewAwsSession(region, credentials)
	if err != nil {
		panic(err)
	}
	manager := assets.NewDefaultFileManager(session)
	endpoints.RegisterAssetsEndpoints(e, manager, bucket)
	e.Logger.Fatal(e.Start(":8080"))
}
