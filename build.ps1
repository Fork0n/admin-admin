# Build script for admin:admin
# Usage: .\build.ps1 [-v "version-name"]
# Example: .\build.ps1 -v "alpha-2.2"

param(
    [Alias("v")]
    [string]$Version
)

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "       admin:admin Build Script        " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# If no version provided via parameter, prompt the user
if ([string]::IsNullOrWhiteSpace($Version)) {
    Write-Host "Enter build version (e.g., alpha-2.2, beta-1.0, release-1.0):" -ForegroundColor Yellow
    $Version = Read-Host "Version"

    # If still empty, use default
    if ([string]::IsNullOrWhiteSpace($Version)) {
        $Version = "dev"
        Write-Host "Using default version: $Version" -ForegroundColor Gray
    }
}

# Sanitize version string (replace spaces and special chars with dashes)
$Version = $Version -replace '[^a-zA-Z0-9\.\-]', '-'

# Build output name
$OutputName = "admin-admin-build-$Version.exe"
$OutputPath = "bin\$OutputName"

Write-Host "Building version: $Version" -ForegroundColor Green
Write-Host "Output file: $OutputPath" -ForegroundColor Green
Write-Host ""

# Enable CGO
$env:CGO_ENABLED = 1

# Create bin directory if it doesn't exist
if (!(Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
    Write-Host "Created bin directory" -ForegroundColor Gray
}

# Build the application
Write-Host "Compiling..." -ForegroundColor Cyan

$buildArgs = @(
    "build",
    "-ldflags=-s -w -X main.Version=$Version",
    "-o", $OutputPath,
    "./cmd/app"
)

go @buildArgs

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "       Build Successful!               " -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Green

    # Show file size
    $fileInfo = Get-Item $OutputPath
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "Size: $sizeMB MB" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "       Build Failed!                   " -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    exit 1
}
