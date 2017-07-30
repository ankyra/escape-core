#!/bin/bash -e

go build 
go test -cover -v $(go list ./... | grep -v -E 'vendor' ) | grep -v "no test files"
