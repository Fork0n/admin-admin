# Run script for admin:admin
# Cross-platform run script for Windows

Write-Host "Running admin:admin..." -ForegroundColor Cyan

# Enable CGO
$env:CGO_ENABLED = 1

# Run the application
go run .\cmd\app

