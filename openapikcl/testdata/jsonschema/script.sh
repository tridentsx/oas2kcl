#!/bin/bash

for file in ./*/dummy.yaml; do
    echo "Processing $file..."

    # Define temporary output file
    temp_output="${file%.yaml}.converted.yaml"

    # Convert JSON to YAML using yq
    if yq -p=json -o=yaml "$file" > "$temp_output"; then
        echo "Converted: $temp_output"
        
        # Rename converted file back to original filename (overwrite)
        mv -f "$temp_output" "$file"
        echo "Renamed: $temp_output → $file"

        # Delete the original JSON file (which is now replaced with YAML content)
        echo "Deleting original JSON file..."
        rm -f "$file"

    else
        echo "Conversion failed for: $file"
        # Cleanup temp output if it was created but failed
        rm -f "$temp_output"
    fi
done

