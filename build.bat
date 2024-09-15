@echo off
REM Check if a version parameter is provided
if "%1"=="" (
    set VERSION=1.0.0
) else (
    set VERSION=%1
)

REM Get the current Git short commit hash
for /f "delims=" %%i in ('git rev-parse --short HEAD') do set GIT_HASH=%%i

REM Append the Git hash to the version
set VERSION=%VERSION%-%GIT_HASH%

echo Building version %VERSION%

REM Build the frontend
cd frontend
npm install
npm run build
cd ..

REM Ensure Go dependencies are up to date
go mod tidy

REM Build Go binaries for different architectures and operating systems
set GOARCH=amd64
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-amd64

set GOARCH=386
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-386

set GOARCH=arm
set GOARM=7
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-arm7

set GOARCH=amd64
set GOOS=windows
go build -o releases\torsniff-%VERSION%-windows-amd64.exe

set GOARCH=386
set GOOS=windows
go build -o releases\torsniff-%VERSION%-windows-386.exe

set GOARCH=amd64
set GOOS=darwin
go build -o releases\torsniff-%VERSION%-darwin-amd64
