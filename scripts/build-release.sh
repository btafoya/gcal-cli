#!/bin/bash
# build-release.sh - Build release binaries for all platforms
# Creates distribution archives with version info

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get version
VERSION=$(./scripts/version.sh current)
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo -e "${GREEN}Building gcal-cli ${VERSION}${NC}"
echo "Commit: ${COMMIT}"
echo "Build Date: ${BUILD_DATE}"
echo ""

# LDFLAGS for version info
LDFLAGS="-X github.com/btafoya/gcal-cli/internal/commands.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X github.com/btafoya/gcal-cli/internal/commands.Commit=${COMMIT}"
LDFLAGS="${LDFLAGS} -X github.com/btafoya/gcal-cli/internal/commands.BuildDate=${BUILD_DATE}"

# Create dist directory
mkdir -p dist

# Build for each platform
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r OS ARCH <<< "$PLATFORM"

    OUTPUT_NAME="gcal-cli-${VERSION}-${OS}-${ARCH}"
    BINARY_NAME="gcal-cli"

    if [[ "$OS" == "windows" ]]; then
        BINARY_NAME="gcal-cli.exe"
    fi

    echo -e "${YELLOW}Building for ${OS}/${ARCH}...${NC}"

    # Build binary
    GOOS=$OS GOARCH=$ARCH go build \
        -ldflags "$LDFLAGS -w -s" \
        -o "dist/${BINARY_NAME}" \
        ./cmd/gcal-cli

    # Create archive
    cd dist
    if [[ "$OS" == "windows" ]]; then
        zip -q "${OUTPUT_NAME}.zip" "${BINARY_NAME}" ../README.md ../LICENSE
        rm "${BINARY_NAME}"
    else
        tar czf "${OUTPUT_NAME}.tar.gz" "${BINARY_NAME}" ../README.md ../LICENSE
        rm "${BINARY_NAME}"
    fi
    cd ..

    echo -e "${GREEN}Created: dist/${OUTPUT_NAME}.tar.gz${NC}"
done

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo "Archives created in dist/ directory:"
ls -lh dist/

# Create checksums
cd dist
sha256sum * > checksums.txt
cd ..

echo ""
echo -e "${GREEN}Checksums created: dist/checksums.txt${NC}"
