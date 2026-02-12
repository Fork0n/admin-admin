#!/bin/bash
# Build script for admin:admin
# Usage: ./build.sh [-v "version-name"]
# Example: ./build.sh -v "alpha-2.2"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}       admin:admin Build Script        ${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# Parse command line arguments
VERSION=""
while getopts "v:" opt; do
    case $opt in
        v)
            VERSION="$OPTARG"
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            exit 1
            ;;
    esac
done

# If no version provided via parameter, prompt the user
if [ -z "$VERSION" ]; then
    echo -e "${YELLOW}Enter build version (e.g., alpha-2.2, beta-1.0, release-1.0):${NC}"
    read -p "Version: " VERSION

    # If still empty, use default
    if [ -z "$VERSION" ]; then
        VERSION="dev"
        echo -e "${GRAY}Using default version: $VERSION${NC}"
    fi
fi

# Sanitize version string (replace spaces and special chars with dashes)
VERSION=$(echo "$VERSION" | sed 's/[^a-zA-Z0-9\.\-]/-/g')

# Detect OS for output extension
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    OUTPUT_NAME="admin-admin-build-$VERSION.exe"
else
    OUTPUT_NAME="admin-admin-build-$VERSION"
fi

OUTPUT_PATH="bin/$OUTPUT_NAME"

echo -e "${GREEN}Building version: $VERSION${NC}"
echo -e "${GREEN}Output file: $OUTPUT_PATH${NC}"
echo ""

# Enable CGO
export CGO_ENABLED=1

# Create bin directory if it doesn't exist
if [ ! -d "bin" ]; then
    mkdir -p bin
    echo -e "${GRAY}Created bin directory${NC}"
fi

# Build the application
echo -e "${CYAN}Compiling...${NC}"

go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_PATH" ./cmd/app

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}       Build Successful!               ${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Output: $OUTPUT_PATH${NC}"

    # Show file size
    if [[ "$OSTYPE" == "darwin"* ]]; then
        SIZE=$(ls -lh "$OUTPUT_PATH" | awk '{print $5}')
    else
        SIZE=$(ls -lh "$OUTPUT_PATH" | awk '{print $5}')
    fi
    echo -e "${GREEN}Size: $SIZE${NC}"
else
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}       Build Failed!                   ${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
