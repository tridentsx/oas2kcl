#!/bin/bash
set -e

echo "Running unit tests..."
go test -v ./openapikcl -run "^Test.*$"

# Check if the converter exists
if [ ! -f "./openapi-to-kcl" ]; then
  echo "Error: openapi-to-kcl binary not found. Build it first with 'go build -o openapi-to-kcl ./cmd/main.go'"
  exit 1
fi

# Create temporary directory for output
temp_dir=$(mktemp -d)
trap "rm -rf $temp_dir" EXIT

overall_result=0

echo "Running integration tests..."
echo "Debug: Input files found:"
ls -la openapikcl/testdata/input/*.json
echo "Debug: Golden directories found:"
ls -la openapikcl/testdata/golden/

# For each OpenAPI file in the testdata/input directory
for input_file in openapikcl/testdata/input/*.json; do
  # Skip non-existent files (in case the glob doesn't match anything)
  [ -f "$input_file" ] || continue
  
  # Extract test name from filename
  test_name=$(basename "${input_file}" .json)
  echo "Testing $test_name..."
  
  # Generate KCL schemas
  ./openapi-to-kcl -oas "${input_file}" -out "${temp_dir}" -package "${test_name}"
  
  # Check if golden directory exists
  golden_dir="openapikcl/testdata/golden/${test_name}"
  if [ ! -d "$golden_dir" ]; then
    echo "⚠️ No golden files found for $test_name, skipping comparison"
    continue
  fi
  
  # Compare each generated file with its golden counterpart
  failed=0
  for generated_file in "$temp_dir"/*.k; do
    filename=$(basename "$generated_file")
    golden_file="$golden_dir/$filename"
    
    if [ ! -f "$golden_file" ]; then
      echo "❌ Golden file missing: $golden_file"
      failed=1
      continue
    fi
    
    if ! diff -u "$golden_file" "$generated_file"; then
      echo "❌ Test failed: $filename doesn't match golden file"
      failed=1
    else
      echo "✅ Test passed: $filename matches golden file"
    fi
  done
  
  if [ $failed -eq 1 ]; then
    overall_result=1
  fi
done

if [ $overall_result -eq 0 ]; then
  echo "All integration tests passed! 🎉"
else
  echo "Some integration tests failed. 😢"
  exit 1
fi
