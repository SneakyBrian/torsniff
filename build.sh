#!/bin/bash

 # Get the latest Git tag for the current branch
 VERSION=$(git describe --tags --abbrev=0 2>/dev/null)

 # If no tag is found, default to 0.0.1
 if [ -z "$VERSION" ]; then
     VERSION="0.0.1"
 fi

# Get the current Git short commit hash
GIT_HASH=$(git rev-parse --short HEAD)

# Append the Git hash to the version
VERSION="${VERSION}-${GIT_HASH}"

echo "Building version ${VERSION}"

# Build the frontend
cd frontend
npm install
npm run build
cd ..

# Ensure Go dependencies are up to date
go mod tidy

# Build Go binaries for different architectures and operating systems
echo "Building Go binary for linux-amd64"
GOARCH=amd64 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-amd64

echo "Building Go binary for linux-arm64"
GOARCH=arm64 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-arm64

echo "Building Go binary for windows-amd64"
GOARCH=amd64 GOOS=windows go build -o releases/torsniff-${VERSION}-windows-amd64.exe

echo "Building Go binary for darwin-amd64"
GOARCH=amd64 GOOS=darwin go build -o releases/torsniff-${VERSION}-darwin-amd64

echo "Build process completed"
