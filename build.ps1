# Build script for Desktop Control System
# Cross-platform build script for Windows

Write-Host "Building Desktop Control System..." -ForegroundColor Cyan

# Enable CGO
$env:CGO_ENABLED = 1

# Determine output binary name
$OUTPUT = "bin\control-system.exe"

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build the application
Write-Host "Compiling..." -ForegroundColor Yellow
go build -v -o $OUTPUT .\cmd\app

# Check if build was successful
if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "Executable created at: $OUTPUT" -ForegroundColor Green
    Write-Host ""
    Write-Host "To run the application:" -ForegroundColor Cyan
    Write-Host "  .\$OUTPUT" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "Build failed!" -ForegroundColor Red
    Write-Host "Make sure you have:" -ForegroundColor Yellow
    Write-Host "  - Go installed and in PATH" -ForegroundColor Yellow
    Write-Host "  - GCC compiler installed (MinGW, MSYS2, or TDM-GCC)" -ForegroundColor Yellow
    Write-Host "  - GCC in your system PATH" -ForegroundColor Yellow
    exit 1
}

