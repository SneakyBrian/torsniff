#!/bin/bash
#!/bin/bash

# Check if a version parameter is provided
if [ -z "$1" ]; then
    VERSION="1.0.0"
else
    VERSION="$1"
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

echo "Building Go binary for linux-386"
GOARCH=386 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-386

echo "Building Go binary for linux-arm7"
GOARCH=arm GOARM=7 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-arm7

echo "Building Go binary for linux-arm64"
GOARCH=arm64 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-arm64

echo "Building Go binary for windows-amd64"
GOARCH=amd64 GOOS=windows go build -o releases/torsniff-${VERSION}-windows-amd64.exe

echo "Building Go binary for windows-386"
GOARCH=386 GOOS=windows go build -o releases/torsniff-${VERSION}-windows-386.exe

echo "Building Go binary for darwin-amd64"
GOARCH=amd64 GOOS=darwin go build -o releases/torsniff-${VERSION}-darwin-amd64

echo "Build process completed"
