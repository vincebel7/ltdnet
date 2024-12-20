#!/bin/bash

# Create the bin directory if it doesn't exist
mkdir -p bin

# Build for Linux
echo "Building ltdnet for Linux..."
GOOS=linux GOARCH=amd64 go build -o bin/ltdnet-linux-v0_5_1 ./src
if [ $? -eq 0 ]; then
    echo "Linux build successful: bin/ltdnet-linux-v0_5_1"
else
    echo "Linux build failed"
    exit 1
fi

# Build for Windows
echo "Building ltdnet for Windows..."
GOOS=windows GOARCH=amd64 go build -o bin/ltdnet-win-v0_5_1.exe ./src
if [ $? -eq 0 ]; then
	echo "Windows build successful: bin/ltdnet-win-v0_5_1.exe"
else
    echo "Windows build failed"
    exit 1
fi

echo "Build complete. Executables are in the bin directory."
