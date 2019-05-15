#!/usr/bin/env bash
set -e
source ./build/checkEnv.sh
go test  ./...
