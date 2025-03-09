#!/bin/bash

# Define the root directory
ROOT_DIR="."

# Find all dummy.yaml files
find "$ROOT_DIR" -type f -name "dummy.yaml" | while read -r file; do
    # Check if the file is valid JSON
    if jq empty "$file" 2>/dev/null; then
        echo "Converting $file..."
        
        # Convert JSON to YAML and overwrite the original file
        jq -r '.' "$file" | yq . > "$file.tmp" && mv "$file.tmp" "$file"
        
        echo "Converted: $file"
    else
        echo "Skipping (not valid JSON): $file"
    fi
done

