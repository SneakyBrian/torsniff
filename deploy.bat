@echo off

REM Example usage: deploy.bat 1.0.0 pi raspberrypi.local /home/pi/

REM Check if all required parameters are provided
if "%1"=="" (
    echo Error: VERSION parameter is missing.
    exit /b 1
)
if "%2"=="" (
    echo Error: REMOTE_USER parameter is missing.
    exit /b 1
)
if "%3"=="" (
    echo Error: REMOTE_HOST parameter is missing.
    exit /b 1
)
if "%4"=="" (
    echo Error: REMOTE_PATH parameter is missing.
    exit /b 1
)

REM Assign parameters to variables
set VERSION=%1
set REMOTE_USER=%2
set REMOTE_HOST=%3
set REMOTE_PATH=%4

REM Define the binary path
set BINARY_PATH=releases\torsniff-%VERSION%-linux-arm64

REM Check if pscp is available
where pscp >nul 2>nul
if errorlevel 1 (
    echo pscp not found. Please ensure PuTTY is installed and pscp is in your PATH.
    exit /b 1
)

REM Copy the binary to the Raspberry Pi
echo Copying %BINARY_PATH% to %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_PATH%
pscp %BINARY_PATH% %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_PATH%

if errorlevel 1 (
    echo Failed to copy the binary to the Raspberry Pi.
    exit /b 1
)

echo Copy completed successfully.
