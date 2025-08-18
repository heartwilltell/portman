#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project info
PROJECT_NAME="wutp"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | cut -d' ' -f3)

echo -e "${BLUE}Building ${PROJECT_NAME} v${VERSION}${NC}"
echo -e "${BLUE}Go version: ${GO_VERSION}${NC}"
echo -e "${BLUE}Build time: ${BUILD_TIME}${NC}"

# Create dist directory
mkdir -p dist

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'main.Version=${VERSION}'"
LDFLAGS="$LDFLAGS -X 'main.BuildTime=${BUILD_TIME}'"

# Function to build for a platform
build_for_platform() {
    local os=$1
    local arch=$2
    local output_name="${PROJECT_NAME}"

    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    local output_path="dist/${PROJECT_NAME}_${os}_${arch}/${output_name}"
    mkdir -p "$(dirname "$output_path")"

    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"

    GOOS=$os GOARCH=$arch go build \
        -ldflags "$LDFLAGS" \
        -o "$output_path" \
        .

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built: $output_path${NC}"
    else
        echo -e "${RED}✗ Failed to build for ${os}/${arch}${NC}"
        exit 1
    fi
}

# Default: build for current platform
if [ $# -eq 0 ]; then
    echo -e "${YELLOW}Building for current platform...${NC}"
    go build -ldflags "$LDFLAGS" -o "dist/${PROJECT_NAME}" .

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built: dist/${PROJECT_NAME}${NC}"
        echo -e "${BLUE}Run with: ./dist/${PROJECT_NAME}${NC}"
    else
        echo -e "${RED}✗ Build failed${NC}"
        exit 1
    fi
fi

# Handle command line arguments
case "${1:-}" in
    "all")
        echo -e "${YELLOW}Building for all platforms...${NC}"
        build_for_platform "linux" "amd64"
        build_for_platform "linux" "arm64"
        build_for_platform "darwin" "amd64"
        build_for_platform "darwin" "arm64"
        build_for_platform "windows" "amd64"
        build_for_platform "windows" "arm64"
        ;;
    "install")
        echo -e "${YELLOW}Building and installing...${NC}"
        go build -ldflags "$LDFLAGS" -o "dist/${PROJECT_NAME}" .

        if [ $? -eq 0 ]; then
            echo -e "${YELLOW}Installing to /usr/local/bin...${NC}"
            sudo cp "dist/${PROJECT_NAME}" "/usr/local/bin/${PROJECT_NAME}"
            sudo chmod +x "/usr/local/bin/${PROJECT_NAME}"
            echo -e "${GREEN}✓ Installed: /usr/local/bin/${PROJECT_NAME}${NC}"
            echo -e "${BLUE}Run with: ${PROJECT_NAME}${NC}"
        else
            echo -e "${RED}✗ Build failed${NC}"
            exit 1
        fi
        ;;
    "clean")
        echo -e "${YELLOW}Cleaning build artifacts...${NC}"
        rm -rf dist/
        echo -e "${GREEN}✓ Cleaned${NC}"
        ;;
    "test")
        echo -e "${YELLOW}Running tests...${NC}"
        go test -v ./...
        ;;
    "fmt")
        echo -e "${YELLOW}Formatting code...${NC}"
        go fmt ./...
        echo -e "${GREEN}✓ Formatted${NC}"
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  (no args)  Build for current platform"
        echo "  all        Build for all supported platforms"
        echo "  install    Build and install to /usr/local/bin"
        echo "  clean      Remove build artifacts"
        echo "  test       Run tests"
        echo "  fmt        Format code"
        echo "  help       Show this help"
        ;;
    *)
        if [ -n "$1" ]; then
            echo -e "${RED}Unknown command: $1${NC}"
            echo "Use '$0 help' for usage information."
            exit 1
        fi
        ;;
esac
