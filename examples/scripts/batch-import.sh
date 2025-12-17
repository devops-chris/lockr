#!/bin/bash
# Batch import secrets from files
# Usage: ./batch-import.sh /path/to/secrets/dir /ssm/prefix

set -e

SECRETS_DIR="${1:?Usage: $0 <secrets-dir> <ssm-prefix>}"
SSM_PREFIX="${2:?Usage: $0 <secrets-dir> <ssm-prefix>}"

# Ensure prefix starts with /
[[ "$SSM_PREFIX" != /* ]] && SSM_PREFIX="/$SSM_PREFIX"

echo "Importing secrets from $SECRETS_DIR to $SSM_PREFIX"
echo ""

for file in "$SECRETS_DIR"/*; do
    if [[ -f "$file" ]]; then
        name=$(basename "$file" | sed 's/\.[^.]*$//')  # Remove extension
        path="${SSM_PREFIX}/${name}"
        
        echo "  → $path"
        lockr write "$path" --file "$file"
    fi
done

echo ""
echo "Done! Imported secrets:"
lockr list "$SSM_PREFIX"

