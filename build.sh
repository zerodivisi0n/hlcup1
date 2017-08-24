#!/bin/sh

set -ex
go get -v .
GOOS=linux GOARCH=amd64 go build -o server .
docker build -t hlcup1 .
