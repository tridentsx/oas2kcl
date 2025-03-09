#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Setup
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

echo "=== Running KCL JSON Schema Tests ==="

run_test() {
    local schema_file=$1
    local valid_data=$2
    local invalid_data=$3
    local schema_name=$4
    
    echo -e "\n=== Testing $schema_name ==="
    
    # Test valid data
    echo "Testing valid data..."
    kcl vet "$valid_data" "$schema_file" -s "$schema_name" &> /dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Valid data passed validation ✓${NC}"
    else
        echo -e "${RED}Valid data failed validation ✗${NC}"
        echo "Command: kcl vet $valid_data $schema_file -s $schema_name"
        kcl vet "$valid_data" "$schema_file" -s "$schema_name"
    fi
    
    # Test invalid data
    echo "Testing invalid data..."
    kcl vet "$invalid_data" "$schema_file" -s "$schema_name" &> /dev/null
    if [ $? -ne 0 ]; then
        echo -e "${GREEN}Invalid data correctly failed validation ✓${NC}"
    else
        echo -e "${RED}Invalid data incorrectly passed validation ✗${NC}"
        echo "Command: kcl vet $invalid_data $schema_file -s $schema_name"
        kcl vet "$invalid_data" "$schema_file" -s "$schema_name"
    fi
}

# Run minimal test
run_test "examples/test_suite/minimal_refactored.k" \
         "examples/test_suite/minimal_valid_data.json" \
         "examples/test_suite/minimal_invalid_data.json" \
         "MinimalValidator"

# Add more test calls here as you develop them

echo -e "\n=== All tests complete ==="

# Run linting on all KCL files
echo -e "\n=== Running KCL linting ==="
find examples -name "*.k" -exec kcl {} \; 2>&1 | grep -v "is empty json" || echo -e "${GREEN}No linting errors found ✓${NC}" 