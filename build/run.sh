#!/usr/bin/env bash
set -e
source ./build/checkEnv.sh
cd cmd/assetuploader
go build -o assetuploader .
chmod a+x assetuploader
./assetuploader --region=${AWS_REGION} --bucket=${AWS_BUCKET}