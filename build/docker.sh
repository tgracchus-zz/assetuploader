#!/usr/bin/env bash
source ./build/distribution.sh

cp ../../distributions/assetuploader-${VERSION}-linux-x86_64 ../../distributions/assetuploader
cd ../../
docker build . -t assetuploader:${VERSION}
