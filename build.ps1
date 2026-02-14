# Build script for admin:admin
# Usage: .\build.ps1 [-Type "dev|release"] [-Stage "alpha|beta|rc"] [-Version "x.y.z"]
# Examples:
#   .\build.ps1                           # Interactive mode
#   .\build.ps1 -Type dev -Version 1.0.1
#   .\build.ps1 -Type release -Stage alpha -Version 1.2.11

param(
    [ValidateSet("dev", "release")]
    [string]$Type,

    [ValidateSet("alpha", "beta", "rc", "")]
    [string]$Stage,

    [string]$Version
)

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "       admin:admin Build Script        " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Version file to track build numbers
$versionFile = ".build-version.json"

# Load or create version tracking
if (Test-Path $versionFile) {
    $versionData = Get-Content $versionFile | ConvertFrom-Json
} else {
    $versionData = @{
        lastType = "dev"
        lastStage = ""
        lastMajor = 1
        lastMinor = 0
        lastBuild = 0
    }
}

# If no type provided, prompt
if ([string]::IsNullOrWhiteSpace($Type)) {
    Write-Host "Select build type:" -ForegroundColor Yellow
    Write-Host "  [1] dev     - Development build"
    Write-Host "  [2] release - Release build"
    $typeChoice = Read-Host "Choice (1/2)"

    switch ($typeChoice) {
        "1" { $Type = "dev" }
        "2" { $Type = "release" }
        default {
            $Type = "dev"
            Write-Host "Using default: dev" -ForegroundColor Gray
        }
    }
}

# If release, ask for stage
if ($Type -eq "release" -and [string]::IsNullOrWhiteSpace($Stage)) {
    Write-Host ""
    Write-Host "Select release stage:" -ForegroundColor Yellow
    Write-Host "  [1] alpha - Alpha release"
    Write-Host "  [2] beta  - Beta release"
    Write-Host "  [3] rc    - Release Candidate"
    Write-Host "  [4] none  - Final release"
    $stageChoice = Read-Host "Choice (1/2/3/4)"

    switch ($stageChoice) {
        "1" { $Stage = "alpha" }
        "2" { $Stage = "beta" }
        "3" { $Stage = "rc" }
        "4" { $Stage = "" }
        default {
            $Stage = "alpha"
            Write-Host "Using default: alpha" -ForegroundColor Gray
        }
    }
}

# Parse or prompt for version
if ([string]::IsNullOrWhiteSpace($Version)) {
    Write-Host ""
    Write-Host "Current version: $($versionData.lastMajor).$($versionData.lastMinor).$($versionData.lastBuild)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Version options:" -ForegroundColor Yellow
    Write-Host "  [1] Auto-increment build number (x.y.z+1)"
    Write-Host "  [2] Increment minor version (x.y+1.0)"
    Write-Host "  [3] Increment major version (x+1.0.0)"
    Write-Host "  [4] Enter custom version"
    $versionChoice = Read-Host "Choice (1/2/3/4)"

    switch ($versionChoice) {
        "1" {
            $major = $versionData.lastMajor
            $minor = $versionData.lastMinor
            $build = $versionData.lastBuild + 1
        }
        "2" {
            $major = $versionData.lastMajor
            $minor = $versionData.lastMinor + 1
            $build = 0
        }
        "3" {
            $major = $versionData.lastMajor + 1
            $minor = 0
            $build = 0
        }
        "4" {
            $customVersion = Read-Host "Enter version (x.y.z)"
            $parts = $customVersion -split '\.'
            $major = [int]($parts[0])
            $minor = [int]($parts[1])
            $build = [int]($parts[2])
        }
        default {
            $major = $versionData.lastMajor
            $minor = $versionData.lastMinor
            $build = $versionData.lastBuild + 1
        }
    }
    $Version = "$major.$minor.$build"
} else {
    # Parse provided version
    $parts = $Version -split '\.'
    $major = [int]$parts[0]
    $minor = if ($parts.Length -gt 1) { [int]$parts[1] } else { 0 }
    $build = if ($parts.Length -gt 2) { [int]$parts[2] } else { 0 }
    $Version = "$major.$minor.$build"
}

# Build the full version string
if ($Type -eq "release" -and -not [string]::IsNullOrWhiteSpace($Stage)) {
    $fullVersion = "$Type $Stage $Version"
    $fileName = "admin-admin-$Type-$Stage-$Version"
} elseif ($Type -eq "release") {
    $fullVersion = "$Type $Version"
    $fileName = "admin-admin-$Type-$Version"
} else {
    $fullVersion = "$Type $Version"
    $fileName = "admin-admin-$Type-$Version"
}

$OutputPath = "bin\$fileName.exe"

Write-Host ""
Write-Host "Building: $fullVersion" -ForegroundColor Green
Write-Host "Output: $OutputPath" -ForegroundColor Green
Write-Host ""

# Enable CGO and add MSYS2 to path
$env:CGO_ENABLED = 1
$env:Path += ";D:\msys2\mingw64\bin"

# Create bin directory if needed
if (!(Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
    Write-Host "Created bin directory" -ForegroundColor Gray
}

# Build
Write-Host "Compiling..." -ForegroundColor Cyan

$buildArgs = @(
    "build",
    "-ldflags=-s -w -X main.Version=$fullVersion",
    "-o", $OutputPath,
    "./cmd/app"
)

go @buildArgs

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "       Build Successful!               " -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "Version: $fullVersion" -ForegroundColor Green
    Write-Host "Output: $OutputPath" -ForegroundColor Green

    $fileInfo = Get-Item $OutputPath
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "Size: $sizeMB MB" -ForegroundColor Green

    # Save version data
    $versionData = @{
        lastType = $Type
        lastStage = $Stage
        lastMajor = $major
        lastMinor = $minor
        lastBuild = $build
    }
    $versionData | ConvertTo-Json | Set-Content $versionFile

    Write-Host ""
    Write-Host "Version saved to $versionFile" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "       Build Failed!                   " -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    exit 1
}
