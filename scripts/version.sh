#!/bin/bash
# version.sh - Semantic versioning script for gcal-cli
# Generates version numbers based on git tags and commit history

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the latest git tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Remove 'v' prefix for version manipulation
VERSION=${LATEST_TAG#v}

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Count commits since last tag
COMMITS_SINCE_TAG=$(git rev-list ${LATEST_TAG}..HEAD --count 2>/dev/null || echo "0")

# Get current commit hash
COMMIT_HASH=$(git rev-parse --short HEAD)

# Check if working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    DIRTY="-dirty"
else
    DIRTY=""
fi

# Determine version based on context
if [[ "$1" == "major" ]]; then
    # Increment major version
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
elif [[ "$1" == "minor" ]]; then
    # Increment minor version
    MINOR=$((MINOR + 1))
    PATCH=0
    NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
elif [[ "$1" == "patch" ]]; then
    # Increment patch version
    PATCH=$((PATCH + 1))
    NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
elif [[ "$1" == "tag" ]]; then
    # Create a new tag
    if [[ -z "$2" ]]; then
        echo -e "${RED}Error: Version number required${NC}"
        echo "Usage: $0 tag <version>"
        echo "Example: $0 tag 1.0.0"
        exit 1
    fi
    NEW_VERSION="$2"

    # Validate version format
    if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}Error: Invalid version format${NC}"
        echo "Version must be in format: MAJOR.MINOR.PATCH (e.g., 1.0.0)"
        exit 1
    fi

    # Create and push tag
    echo -e "${GREEN}Creating tag v${NEW_VERSION}${NC}"
    git tag -a "v${NEW_VERSION}" -m "Release v${NEW_VERSION}"
    echo -e "${YELLOW}Tag created. Push with: git push origin v${NEW_VERSION}${NC}"
    exit 0
elif [[ "$1" == "current" ]] || [[ -z "$1" ]]; then
    # Show current version
    if [[ $COMMITS_SINCE_TAG -gt 0 ]]; then
        # Development version
        NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}-dev.${COMMITS_SINCE_TAG}+${COMMIT_HASH}${DIRTY}"
    else
        # Released version
        NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}${DIRTY}"
    fi
elif [[ "$1" == "next" ]]; then
    # Show next patch version
    PATCH=$((PATCH + 1))
    NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
elif [[ "$1" == "help" ]] || [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    echo "Usage: $0 [command] [args]"
    echo ""
    echo "Commands:"
    echo "  current       Show current version (default)"
    echo "  next          Show next patch version"
    echo "  major         Increment major version (X.0.0)"
    echo "  minor         Increment minor version (0.X.0)"
    echo "  patch         Increment patch version (0.0.X)"
    echo "  tag <version> Create and tag a new version"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Show current version"
    echo "  $0 current           # Show current version"
    echo "  $0 next              # Show next patch version"
    echo "  $0 tag 1.0.0         # Create v1.0.0 tag"
    echo ""
    echo "Version format: MAJOR.MINOR.PATCH[-dev.N+HASH]"
    echo "  - Released versions: 1.0.0"
    echo "  - Development versions: 1.0.0-dev.5+a1b2c3d"
    exit 0
else
    echo -e "${RED}Error: Unknown command '$1'${NC}"
    echo "Run '$0 help' for usage information"
    exit 1
fi

echo "$NEW_VERSION"
