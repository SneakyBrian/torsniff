@echo off
 REM Get the latest Git tag for the current branch
 for /f "delims=" %%i in ('git describe --tags --abbrev=0 2^>nul') do set VERSION=%%i

 REM If no tag is found, default to 1.0.0
 if "%VERSION%"=="" (
     set VERSION=1.0.0
 )

REM Get the current Git short commit hash
for /f "delims=" %%i in ('git rev-parse --short HEAD') do set GIT_HASH=%%i

REM Append the Git hash to the version
set VERSION=%VERSION%-%GIT_HASH%

echo Building version %VERSION%

REM Build the frontend
cd frontend
call npm install
call npm run build
cd ..

REM Ensure Go dependencies are up to date
go mod tidy

REM Build Go binaries for different architectures and operating systems
echo Building Go binary for linux-amd64
set GOARCH=amd64
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-amd64

echo Building Go binary for linux-386
set GOARCH=386
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-386

echo Building Go binary for linux-arm7
set GOARCH=arm
set GOARM=7
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-arm7

echo Building Go binary for linux-arm64
set GOARCH=arm64
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-arm64

echo Building Go binary for windows-amd64
set GOARCH=amd64
set GOOS=windows
go build -o releases\torsniff-%VERSION%-windows-amd64.exe

echo Building Go binary for windows-386
set GOARCH=386
set GOOS=windows
go build -o releases\torsniff-%VERSION%-windows-386.exe

echo Building Go binary for darwin-amd64
set GOARCH=amd64
set GOOS=darwin
go build -o releases\torsniff-%VERSION%-darwin-amd64

echo Build process completed
