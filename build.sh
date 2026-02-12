#!/bin/bash
fi
    exit 1
    echo "Build failed!"
    echo ""
else
    chmod +x "$OUTPUT"
    echo "  ./$OUTPUT"
    echo "To run the application:"
    echo ""
    echo "Executable created at: $OUTPUT"
    echo "Build successful!"
    echo ""
if [ $? -eq 0 ]; then
# Check if build was successful

go build -v -o "$OUTPUT" ./cmd/app
# Build the application

mkdir -p bin
# Create bin directory if it doesn't exist

fi
    echo "Building for Linux..."
    OUTPUT="bin/control-system"
else
    echo "Building for macOS..."
    OUTPUT="bin/control-system"
if [[ "$OSTYPE" == "darwin"* ]]; then
# Determine output binary name based on OS

export CGO_ENABLED=1
# Enable CGO

echo "Building Desktop Control System..."

# Cross-platform build script for Linux/Mac
# Build script for Desktop Control System

