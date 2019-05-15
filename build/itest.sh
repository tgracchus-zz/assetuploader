#!/usr/bin/env bash
set -e
if [[ -z "${AWS_ACCESS_KEY_ID}" ]]; then
    echo "Please, provide the env var AWS_ACCESS_KEY_ID to run test"
    exit 1
fi

if [[ -z "${AWS_SECRET_ACCESS_KEY}" ]]; then
    echo "Please, provide the env var AWS_SECRET_ACCESS_KEY to run test"
    exit 1
fi

if [[ -z "${TEST_REGION}" ]]; then
    echo "Please, provide the env var TEST_REGION to run test"
    exit 1
fi

if [[ -z "${TEST_BUCKET}" ]]; then
    echo "Please, provide the env var TEST_BUCKET to run test"
    exit 1
fi
go test  ./...
