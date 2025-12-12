#!/bin/bash
# Sync version between VERSION file and version/version.go
#
# Usage:
#   ./sync_version.sh          # Sync from VERSION to version.go
#   ./sync_version.sh 0.7.3    # Set new version in both files
#   ./sync_version.sh --check  # Check if they match (for CI)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

VERSION_FILE="VERSION"
VERSION_GO="version/version.go"

get_file_version() {
    cat "$VERSION_FILE" | tr -d '[:space:]'
}

get_go_version() {
    grep 'const Version' "$VERSION_GO" | sed 's/.*"\(.*\)".*/\1/'
}

update_go_version() {
    local ver="$1"
    sed -i.bak "s/const Version = \".*\"/const Version = \"$ver\"/" "$VERSION_GO"
    rm -f "${VERSION_GO}.bak"
}

update_file_version() {
    local ver="$1"
    echo "$ver" > "$VERSION_FILE"
}

case "${1:-sync}" in
    --check|-c)
        file_ver=$(get_file_version)
        go_ver=$(get_go_version)
        if [ "$file_ver" = "$go_ver" ]; then
            echo "Versions match: $file_ver"
            exit 0
        else
            echo "Version mismatch!"
            echo "  VERSION:          $file_ver"
            echo "  version/version.go: $go_ver"
            exit 1
        fi
        ;;
    sync)
        file_ver=$(get_file_version)
        go_ver=$(get_go_version)
        if [ "$file_ver" = "$go_ver" ]; then
            echo "Versions already in sync: $file_ver"
        else
            update_go_version "$file_ver"
            echo "Updated version/version.go: $go_ver -> $file_ver"
        fi
        ;;
    *)
        # Assume it's a version number
        new_ver="$1"
        if [[ ! "$new_ver" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "Invalid version format: $new_ver"
            echo "Expected: X.Y.Z (e.g., 0.7.3)"
            exit 1
        fi
        update_file_version "$new_ver"
        update_go_version "$new_ver"
        echo "Set version to $new_ver in both files"
        ;;
esac
