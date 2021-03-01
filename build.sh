#!/bin/bash
GOARCH=amd64 GOOS=linux go build -o releases/torsniff-search-${VERSION}-linux-amd64
GOARCH=386 GOOS=linux go build -o releases/torsniff-search-${VERSION}-linux-386

GOARCH=amd64 GOOS=windows go build -o releases/torsniff-search-${VERSION}-windows-amd64.exe
GOARCH=386 GOOS=windows go build -o releases/torsniff-search-${VERSION}-windows-386.exe

GOARCH=amd64 GOOS=darwin go build -o releases/torsniff-search-${VERSION}-darwin-amd64
