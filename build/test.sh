#!/usr/bin/env bash
set -e
go test -cover -race `go list ./... | grep -v assets`
