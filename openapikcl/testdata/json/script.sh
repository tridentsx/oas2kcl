#!/bin/bash

for file in ./*/dummy.yaml; do
    echo "Processing $file..."

    # Get the directory where the file is located
    dir=$(dirname "$file")

    # Define temporary output file inside the same directory
    temp_output="${dir}/dummy.converted.yaml"

    # Convert JSON to YAML using yq
    if yq -p=json -o=yaml "$file" > "$temp_output"; then
        echo "Converted: $temp_output"
        
        # Rename converted file back to dummy.yaml (overwrite original)
        mv -f "$temp_output" "$file"
        echo "Renamed: $temp_output → $file"

        # Delete the original JSON file (which is now replaced with YAML content)
        echo "Deleting original JSON file: $file"
        rm -f "$file"

    else
        echo "Conversion failed for: $file"
        # Cleanup temp output if it was created but failed
        rm -f "$temp_output"
    fi
done

