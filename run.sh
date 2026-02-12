#!/bin/bash
# Run script for admin:admin
# Cross-platform run script for Linux/Mac

echo "Running admin:admin..."

# Enable CGO
export CGO_ENABLED=1

# Run the application
go run ./cmd/app

