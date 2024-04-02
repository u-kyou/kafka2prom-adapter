#!/usr/bin/env bash

set -o errexit
set -o nounset

OUTPUT_DIR=bin
BUILDPATH=./${1:?"need arg: specify a path to build"}
OUTPUT_NAME=${OUTPUT_DIR}/"kafka2prom-adapter"

BUILD_GOOS=${GOOS:-$(go env GOOS)}
BUILD_GOARCH=${GOARCH:-$(go env GOARCH)}
GOBINARY=${GOBINARY:-go}

COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_DATE=$(date "+%F %T")
VERSION=$(git describe --tags --always)

LDFLAGS="-X 'kafka2prom-adapter/pkg/config.Version=$VERSION' -X 'kafka2prom-adapter/pkg/config.Commit=$COMMIT_HASH' -X 'kafka2prom-adapter/pkg/config.BuildDate=$BUILD_DATE'" 

GOOS=${BUILD_GOOS} CGO_ENABLED=1 GOARCH=${BUILD_GOARCH} ${GOBINARY} build \
        -ldflags="${LDFLAGS}" \
        -o "${OUTPUT_NAME}" \
        "${BUILDPATH}"