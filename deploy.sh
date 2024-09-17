#!/bin/bash

# Example usage: ./deploy.sh 1.0.0 pi raspberrypi.local /home/pi/

# Check if all required parameters are provided
if [ -z "$1" ]; then
    echo "Error: VERSION parameter is missing."
    exit 1
fi
if [ -z "$2" ]; then
    echo "Error: REMOTE_USER parameter is missing."
    exit 1
fi
if [ -z "$3" ]; then
    echo "Error: REMOTE_HOST parameter is missing."
    exit 1
fi
if [ -z "$4" ]; then
    echo "Error: REMOTE_PATH parameter is missing."
    exit 1
fi

# Assign parameters to variables
VERSION=$1
REMOTE_USER=$2
REMOTE_HOST=$3
REMOTE_PATH=$4

# Define the binary path
BINARY_PATH="releases/torsniff-${VERSION}-linux-arm64"

# Check if scp is available
if ! command -v scp &> /dev/null; then
    echo "scp not found. Please ensure OpenSSH is installed and scp is in your PATH."
    exit 1
fi

# Copy the binary to the Raspberry Pi
echo "Copying ${BINARY_PATH} to ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}"
scp "${BINARY_PATH}" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}"

if [ $? -ne 0 ]; then
    echo "Failed to copy the binary to the Raspberry Pi."
    exit 1
fi

echo "Copy completed successfully."
