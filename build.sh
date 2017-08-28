#!/bin/sh

set -ex
go get -v .
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -extldflags "-static"' -o server
docker build -t hlcup1 .
