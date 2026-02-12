# Run script for Desktop Control System
# Cross-platform run script for Windows

Write-Host "Running Desktop Control System..." -ForegroundColor Cyan

# Enable CGO
$env:CGO_ENABLED = 1

# Run the application
go run .\cmd\app

