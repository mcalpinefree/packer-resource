#!/bin/bash
set -e
export CGO_ENABLED=0
BUILD_DIR=$(pwd)
ARTIFACT=$1
mkdir -p /go/src/github.com/ci-pipeline/
cp -r resource /go/src/github.com/ci-pipeline/
cd /go/src/github.com/ci-pipeline/resource/${ARTIFACT}
go get
go test
go build
cp ${ARTIFACT} ${BUILD_DIR}/${ARTIFACT}/${ARTIFACT}
