#!/bin/bash
# admin:admin Build Script
# Usage: ./build.sh [-s] [-v version] [-o w|l|m]

silent=false
version=""
os=""

# Parse arguments
while getopts "sv:o:" opt; do
    case $opt in
        s) silent=true ;;
        v) version="$OPTARG" ;;
        o) os="$OPTARG" ;;
    esac
done

log() {
    if [ "$silent" = false ]; then echo "$1"; fi
}

# Detect current OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "l" ;;
        Darwin*) echo "m" ;;
        MINGW*|CYGWIN*|MSYS*) echo "w" ;;
        *) echo "l" ;;
    esac
}
currentOS=$(detect_os)

# If no OS specified, ask
if [ -z "$os" ]; then
    echo "Select target OS:"
    echo "  [w] Windows"
    echo "  [l] Linux"
    echo "  [m] macOS"
    echo "(Note: Cross-compilation requires native OS or Docker)"
    read -p "OS (w/l/m): " os
fi

# Validate OS
case "$os" in
    w|W) export GOOS=windows; ext=".exe"; targetOS="w" ;;
    l|L) export GOOS=linux; ext=""; targetOS="l" ;;
    m|M) export GOOS=darwin; ext=""; targetOS="m" ;;
    *)
        echo "Invalid OS. Build aborted."
        exit 1
        ;;
esac

# Warn about cross-compilation
if [ "$targetOS" != "$currentOS" ]; then
    echo "Warning: Cross-compiling Fyne apps requires native build environment."
    echo "CGO is required for Fyne. Build may fail."
fi

# If no version, ask
if [ -z "$version" ]; then
    read -p "Version (leave empty for none): " version
fi

# Build filename
filename="admin-admin"
[ -n "$version" ] && filename="$filename-$version"
filename="$filename$ext"
output="bin/$filename"

# Setup environment
export CGO_ENABLED=1
export GOARCH=amd64

# Create bin directory
mkdir -p bin

log "Building $filename..."

# Build
if go build -ldflags="-s -w" -o "$output" ./cmd/app 2>&1; then
    log "Build successful: $output"
    if [ "$silent" = false ]; then
        size=$(du -h "$output" | cut -f1)
        echo "Size: $size"
    fi
else
    echo "Build failed!"
    [ "$targetOS" != "$currentOS" ] && echo "Cross-compilation failed. Build on target OS instead."
    exit 1
fi
