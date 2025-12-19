#!/bin/bash
# Create release zip excluding .git, .gitignore matches, and @-prefixed files

set -e

VERSION="${1:-0.7.4}"
OUTPUT="ual-${VERSION}.zip"

if [ ! -f ".gitignore" ]; then
    echo "Error: Run from project root (no .gitignore found)"
    exit 1
fi

echo "Creating $OUTPUT..."

# Create temp file with exclusion patterns
EXCLUDE_FILE=$(mktemp)
trap "rm -f $EXCLUDE_FILE" EXIT

# Core exclusions
cat >> "$EXCLUDE_FILE" << 'PATTERNS'
.git/*
.git
.gitignore
PATTERNS

# Add @-prefixed files (list actual files, not glob)
find . -name '@*' -print | sed 's|^\./||' >> "$EXCLUDE_FILE"

# Add patterns from .gitignore
while IFS= read -r line; do
    # Skip empty lines and comments
    [[ -z "$line" || "$line" =~ ^# ]] && continue
    # Trim whitespace
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -z "$line" ]] && continue
    echo "$line" >> "$EXCLUDE_FILE"
done < .gitignore

# Create zip
rm -f "$OUTPUT"
zip -r "$OUTPUT" . -x@"$EXCLUDE_FILE"

echo ""
echo "Created: $OUTPUT"
ls -lh "$OUTPUT"