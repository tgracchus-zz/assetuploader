#!/usr/bin/env bash
set -e
if [[ -z "${AWS_ACCESS_KEY_ID}" ]]; then
    echo "Please, provide the env var AWS_ACCESS_KEY_ID"
    exit 1
fi

if [[ -z "${AWS_SECRET_ACCESS_KEY}" ]]; then
    echo "Please, provide the env var AWS_SECRET_ACCESS_KEY"
    exit 1
fi

if [[ -z "${AWS_REGION}" ]]; then
    echo "Please, provide the env var AWS_REGION"
    exit 1
fi

if [[ -z "${AWS_BUCKET}" ]]; then
    echo "Please, provide the env var AWS_BUCKET"
    exit 1
fi