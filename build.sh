#!/bin/bash
# Build the frontend
cd frontend
npm install
npm run build
cd ..

# Ensure Go dependencies are up to date
go mod tidy

GOARCH=amd64 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-amd64
GOARCH=386 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-386
GOARCH=arm GOARM=7 GOOS=linux go build -o releases/torsniff-${VERSION}-linux-arm7

GOARCH=amd64 GOOS=windows go build -o releases/torsniff-${VERSION}-windows-amd64.exe
GOARCH=386 GOOS=windows go build -o releases/torsniff-${VERSION}-windows-386.exe

GOARCH=amd64 GOOS=darwin go build -o releases/torsniff-${VERSION}-darwin-amd64
