#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Setup
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

echo -e "${YELLOW}=== Cleaning up unnecessary files and folders ===${NC}"

# List of directories to clean up
CLEANUP_DIRS=(
    "examples/output"
    "examples/output-arrays"
    "examples/output-final"
    "examples/output-fixed"
    "examples/output-new"
    "examples/output-numbers"
    "examples/output-templates"
    "examples/output-validation1"
    "examples/output-validation2"
    "examples/output-validation3"
    "examples/output-validation4"
    "examples/output-validation5"
)

# Remove each directory
for dir in "${CLEANUP_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo -e "Removing directory: ${YELLOW}$dir${NC}"
        rm -rf "$dir"
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}Successfully removed $dir${NC}"
        else
            echo -e "${RED}Failed to remove $dir${NC}"
        fi
    else
        echo -e "Directory ${YELLOW}$dir${NC} does not exist, skipping"
    fi
done

# Keep only necessary test files
echo -e "\n${YELLOW}=== Organizing test files ===${NC}"

# Create a clean test directory
mkdir -p examples/test_suite/clean
cp examples/test_suite/comprehensive_test.json examples/test_suite/clean/
cp examples/test_suite/minimal_valid_data.json examples/test_suite/clean/
cp examples/test_suite/minimal_invalid_data.json examples/test_suite/clean/
cp examples/test_suite/minimal_refactored.k examples/test_suite/clean/

echo -e "${GREEN}Cleanup completed successfully!${NC}"
echo -e "The following files and directories have been preserved:"
echo -e "- ${YELLOW}examples/string_constraints.json${NC} (String constraint test)"
echo -e "- ${YELLOW}examples/number_constraints.json${NC} (Number constraint test)"
echo -e "- ${YELLOW}examples/array_constraints.json${NC} (Array constraint test)"
echo -e "- ${YELLOW}examples/test_suite/clean/${NC} (Clean test files)"
echo -e "- ${YELLOW}examples/test_suite/comprehensive_test.json${NC} (Comprehensive test schema)" 