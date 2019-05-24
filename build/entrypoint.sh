#!/usr/bin/env bash
set -e
source ./bin/checkEnv.sh
echo ${AWS_REGION}
echo ${AWS_BUCKET}
echo $AWS_REGION
echo $AWS_BUCKET
./bin/assetuploader --region=${AWS_REGION} --bucket=${AWS_BUCKET}