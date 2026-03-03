#!/usr/bin/env bash

set -e

# Configuration
APP_NAME="squix"
VERSION=${1:-"dev"}
BUILD_DIR="./dist"
DOCKER_IMAGE="ghcr.io/goreleaser/goreleaser-cross:v1.25"

echo "Building $APP_NAME version $VERSION"
echo "=================================="

# Clean previous builds
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Function to build in Docker
build_with_docker() {
    local platform=$1
    local goos=$2
    local goarch=$3
    local cc=$4
    local output=$5
    
    echo ""
    echo "Building $platform..."
    
    docker run --rm \
        --entrypoint="" \
        -v "$(pwd):/workspace" \
        -w /workspace \
        -e CGO_ENABLED=1 \
        -e GOOS=$goos \
        -e GOARCH=$goarch \
        -e CC=$cc \
        $DOCKER_IMAGE \
        bash -c "git config --global --add safe.directory /workspace && go build -buildvcs=false -ldflags='-s -w' -o $output ./cmd/squix"
    
    echo "✓ Built $platform successfully"
}

# Build Linux AMD64
build_with_docker "Linux AMD64" "linux" "amd64" "gcc" "$BUILD_DIR/${APP_NAME}-linux-amd64"

# Build Linux ARM64
build_with_docker "Linux ARM64" "linux" "arm64" "aarch64-linux-gnu-gcc" "$BUILD_DIR/${APP_NAME}-linux-arm64"

# Build Windows AMD64
build_with_docker "Windows AMD64" "windows" "amd64" "x86_64-w64-mingw32-gcc" "$BUILD_DIR/${APP_NAME}-windows-amd64.exe"

# Build macOS binaries without CGO (Oracle/SQLite will not be available)
echo ""
echo "Building macOS AMD64 (without CGO - Oracle/SQLite unavailable)..."
docker run --rm \
    --entrypoint="" \
    -v "$(pwd):/workspace" \
    -w /workspace \
    -e CGO_ENABLED=0 \
    -e GOOS=darwin \
    -e GOARCH=amd64 \
    $DOCKER_IMAGE \
    bash -c "git config --global --add safe.directory /workspace && go build -buildvcs=false -ldflags='-s -w' -tags=nocgo -o $BUILD_DIR/${APP_NAME}-darwin-amd64 ./cmd/squix"
echo "✓ Built macOS AMD64 successfully"

echo ""
echo "Building macOS ARM64 (without CGO - Oracle/SQLite unavailable)..."
docker run --rm \
    --entrypoint="" \
    -v "$(pwd):/workspace" \
    -w /workspace \
    -e CGO_ENABLED=0 \
    -e GOOS=darwin \
    -e GOARCH=arm64 \
    $DOCKER_IMAGE \
    bash -c "git config --global --add safe.directory /workspace && go build -buildvcs=false -ldflags='-s -w' -tags=nocgo -o $BUILD_DIR/${APP_NAME}-darwin-arm64 ./cmd/squix"
echo "✓ Built macOS ARM64 successfully"

echo ""
echo "=================================="
echo "Build complete!"
echo "=================================="
echo ""
echo "Platforms built:"
echo "  ✓ Linux AMD64 (with CGO - all drivers available)"
echo "  ✓ Linux ARM64 (with CGO - all drivers available)"
echo "  ✓ Windows AMD64 (with CGO - all drivers available)"
echo "  ✓ macOS AMD64 (without CGO - Oracle/SQLite unavailable)"
echo "  ✓ macOS ARM64 (without CGO - Oracle/SQLite unavailable)"
echo ""
echo "Note: macOS binaries will show errors if Oracle/SQLite are used."
echo "For full macOS support with CGO, build on a Mac with:"
echo "  CGO_ENABLED=1 go build -o squix ./cmd/squix"
echo ""

# Create source code archives
echo "Creating source code archives..."
mkdir -p "$BUILD_DIR/source"
git archive --format=tar.gz --prefix="${APP_NAME}-${VERSION}/" -o "$BUILD_DIR/source/${APP_NAME}_${VERSION}_source.tar.gz" HEAD
git archive --format=zip --prefix="${APP_NAME}-${VERSION}/" -o "$BUILD_DIR/source/${APP_NAME}_${VERSION}_source.zip" HEAD
echo "✓ Source archives created"

echo ""
echo "Generating checksums for all artifacts..."
find "$BUILD_DIR" -type f -exec sha256sum {} + | sed "s|$BUILD_DIR/||" > "$BUILD_DIR/checksums.txt"
echo "✓ Checksums generated in $BUILD_DIR/checksums.txt"

echo ""
echo "=================================="
echo "All tasks complete!"
echo "=================================="
echo ""
echo "Binaries location: $BUILD_DIR/"
echo "Source archives location: $BUILD_DIR/source/"
echo "Checksum file: $BUILD_DIR/checksums.txt"
echo ""
ls -lh "$BUILD_DIR/"
echo ""
ls -lh "$BUILD_DIR/source/"
echo ""
