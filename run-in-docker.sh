#!/bin/sh

CURRENT=$(cd "$(dirname "$0")" && pwd)
docker volume create mackerel-cloudwatch-forwarder-cache > /dev/null 2>&1
docker run --rm -it \
    -e GO111MODULE=on \
    -v mackerel-cloudwatch-forwarder-cache:/go/pkg/mod \
    -v "$CURRENT":/go/src/github.com/shogo82148/mackerel-cloudwatch-forwarder \
    -w /go/src/github.com/shogo82148/mackerel-cloudwatch-forwarder golang:1.18.2 "$@"
