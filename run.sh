#!/bin/bash
# Run script for Desktop Control System
# Cross-platform run script for Linux/Mac

echo "Running Desktop Control System..."

# Enable CGO
export CGO_ENABLED=1

# Run the application
go run ./cmd/app

