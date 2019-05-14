#!/usr/bin/env bash
go test -v ./...

if [[ -z "${AWS_ACCESS_KEY_ID}" ]]; then
    echo "Please, provide the env var AWS_ACCESS_KEY_ID to run test"
    exit(1)
fi