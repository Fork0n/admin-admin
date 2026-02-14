# admin:admin Build Script
# Usage: .\build.ps1 [-s] [-v version] [-o w|l|m]
param(
    [switch]$s,           # Silent mode
    [string]$v = "",      # Version
    [string]$o = ""       # OS: w=windows, l=linux, m=mac
)
$silent = $s
function Log($msg) {
    if (-not $silent) { Write-Host $msg }
}
# Detect current OS
$currentOS = if ($env:OS -eq "Windows_NT") { "w" } elseif ($IsMacOS) { "m" } else { "l" }
# If no OS specified, ask interactively
if ($o -eq "") {
    Write-Host "Select target OS:"
    Write-Host "  [w] Windows"
    Write-Host "  [l] Linux"
    Write-Host "  [m] macOS"
    Write-Host "(Note: Cross-compilation requires native OS or Docker)" -ForegroundColor Yellow
    $o = Read-Host "OS (w/l/m)"
}
# Validate OS
switch ($o.ToLower()) {
    "w" { $env:GOOS = "windows"; $ext = ".exe"; $targetOS = "w" }
    "l" { $env:GOOS = "linux"; $ext = ""; $targetOS = "l" }
    "m" { $env:GOOS = "darwin"; $ext = ""; $targetOS = "m" }
    default {
        Write-Host "Invalid OS. Build aborted." -ForegroundColor Red
        exit 1
    }
}
# Warn about cross-compilation
if ($targetOS -ne $currentOS) {
    Write-Host "Warning: Cross-compiling Fyne apps requires native build environment." -ForegroundColor Yellow
    Write-Host "CGO is required for Fyne. Build may fail." -ForegroundColor Yellow
    Write-Host "Consider using Docker or building on target OS." -ForegroundColor Yellow
}
# If no version, ask
if ($v -eq "") {
    $v = Read-Host "Version (leave empty for none)"
}
# Build filename
$filename = "admin-admin"
if ($v -ne "") { $filename += "-$v" }
$filename += $ext
$output = "bin\$filename"
# Setup environment
$env:CGO_ENABLED = "1"
$env:GOARCH = "amd64"
$env:Path += ";D:\msys2\mingw64\bin"
# Create bin directory
if (!(Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
Log "Building $filename..."
# Build
$buildOutput = go build -ldflags="-s -w" -o $output ./cmd/app 2>&1
$buildSuccess = $LASTEXITCODE -eq 0
if (-not $silent) {
    $buildOutput | ForEach-Object { Write-Host $_ }
}
if ($buildSuccess) {
    Log "Build successful: $output"
    if (-not $silent) {
        $size = [math]::Round((Get-Item $output).Length / 1MB, 2)
        Write-Host "Size: $size MB" -ForegroundColor Green
    }
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    if ($targetOS -ne $currentOS) {
        Write-Host "Cross-compilation failed. Build on target OS instead." -ForegroundColor Yellow
    }
    exit 1
}
