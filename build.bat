@echo off

setlocal enabledelayedexpansion

REM Get the latest Git tag for the current branch
for /f "delims=" %%i in ('git describe --tags --abbrev=0 2^>nul') do set LATEST_TAG=%%i

@REM echo Latest Git Tag: %LATEST_TAG%

REM If no tag is found, default to 0.0.1
if "%LATEST_TAG%"=="" (
    set LATEST_TAG=0.0.1
    @REM echo Defaulting to: %LATEST_TAG%
)

REM Get the current Git short commit hash
for /f "delims=" %%i in ('git rev-parse --short HEAD') do set GIT_HASH=%%i

@REM echo Git Hash: %GIT_HASH%

REM Append the Git hash to the version
set VERSION=%LATEST_TAG%-%GIT_HASH%

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

echo Building Go binary for linux-arm64
set GOARCH=arm64
set GOOS=linux
go build -o releases\torsniff-%VERSION%-linux-arm64

echo Building Go binary for windows-amd64
set GOARCH=amd64
set GOOS=windows
go build -o releases\torsniff-%VERSION%-windows-amd64.exe

echo Building Go binary for darwin-amd64
set GOARCH=amd64
set GOOS=darwin
go build -o releases\torsniff-%VERSION%-darwin-amd64

echo Build process completed
