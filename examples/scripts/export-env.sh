#!/bin/bash
# Export secrets as environment variables
# Usage: eval $(./export-env.sh /myapp/prod)

set -e

SSM_PATH="${1:?Usage: $0 <ssm-path>}"

# Get all secrets at path and export as env vars
lockr list "$SSM_PATH" --recursive --output json | jq -r '
    .[] | 
    .name | 
    split("/") | 
    last | 
    ascii_upcase | 
    gsub("-"; "_")
' | while read -r var_name; do
    # This is just an example - in practice you'd need the actual path
    echo "# export $var_name=\$(lockr read <path> -q)"
done

echo ""
echo "# Example usage:"
echo "# export DB_PASSWORD=\$(lockr read /myapp/prod/db-password -q)"
echo "# export API_KEY=\$(lockr read /myapp/prod/api-key -q)"

