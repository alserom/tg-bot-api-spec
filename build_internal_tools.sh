#!/usr/bin/env bash

COMMIT_HASH="$(git rev-parse --short HEAD)"
BUILD_TIMESTAMP=$(date '+%Y-%m-%dT%H:%M:%S')
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64

LDFLAGS=(
  "-X 'main.CommitHash=${COMMIT_HASH}'"
  "-X 'main.BuildDate=${BUILD_TIMESTAMP}'"
)

TOOLS=(
  validate-spec
  to-repo-data
)

for TOOL in "${TOOLS[@]}"
do
	env CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$GOARCH go build -o .bin/$TOOL -ldflags="${LDFLAGS[*]}" cmd/$TOOL/main.go
done
