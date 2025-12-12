#!/bin/bash
# Renumber all examples with three-digit sequential numbering
# Fills gaps, assigns numbers to unnumbered files
# Compatible with bash 3.2+ (macOS default)

set -e

EXAMPLES_DIR="${1:-examples}"

if [ ! -d "$EXAMPLES_DIR" ]; then
    echo "Error: Directory '$EXAMPLES_DIR' not found"
    exit 1
fi

cd "$EXAMPLES_DIR"

# Collect all .ual files into temp files
numbered_tmp=$(mktemp)
unnumbered_tmp=$(mktemp)
trap "rm -f $numbered_tmp $unnumbered_tmp" EXIT

for f in *.ual; do
    if echo "$f" | grep -qE '^[0-9]+_'; then
        echo "$f" >> "$numbered_tmp"
    else
        echo "$f" >> "$unnumbered_tmp"
    fi
done

# Sort numbered files numerically, unnumbered alphabetically
sort -t_ -k1 -n "$numbered_tmp" > "${numbered_tmp}.sorted"
sort "$unnumbered_tmp" > "${unnumbered_tmp}.sorted"

# Combine into one list
all_files_tmp=$(mktemp)
cat "${numbered_tmp}.sorted" "${unnumbered_tmp}.sorted" > "$all_files_tmp"

total=$(wc -l < "$all_files_tmp" | tr -d ' ')
echo "Found $total example files"
echo ""
echo "Renumbering plan:"
echo "================="

# Build rename plan
plan_tmp=$(mktemp)
counter=1

while IFS= read -r f; do
    # Extract the base name (strip any existing number prefix)
    if echo "$f" | grep -qE '^[0-9]+_'; then
        base=$(echo "$f" | sed 's/^[0-9]*_//')
    else
        base="$f"
    fi
    
    # Generate new name with 3-digit prefix
    new_name=$(printf "%03d_%s" $counter "$base")
    
    if [ "$f" != "$new_name" ]; then
        echo "  $f -> $new_name"
        echo "$f|$new_name" >> "$plan_tmp"
    else
        echo "  $f (unchanged)"
    fi
    
    counter=$((counter + 1))
done < "$all_files_tmp"

to_rename=$(wc -l < "$plan_tmp" | tr -d ' ')
echo ""
echo "Total: $total files, $to_rename to rename"
echo ""

# Ask for confirmation
printf "Proceed with renaming? [y/N] "
read confirm
case "$confirm" in
    [Yy]*) ;;
    *) echo "Aborted."; exit 0 ;;
esac

# Rename files using temp suffix to avoid collisions
echo ""
echo "Renaming..."

# First pass: rename to temp names
while IFS='|' read -r old new; do
    mv "$old" "${new}.tmp"
done < "$plan_tmp"

# Second pass: remove .tmp suffix
while IFS='|' read -r old new; do
    mv "${new}.tmp" "$new"
done < "$plan_tmp"

echo "Done. Renamed $to_rename files."
echo ""
echo "New listing:"
ls -1 *.ual | head -20
echo "..."
ls -1 *.ual | tail -10

# Cleanup
rm -f "${numbered_tmp}.sorted" "${unnumbered_tmp}.sorted" "$all_files_tmp" "$plan_tmp"
