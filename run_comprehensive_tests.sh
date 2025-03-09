#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Setup
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

echo "=== Running Comprehensive KCL JSON Schema Tests ==="

# Clean up old test directories
cleanup() {
    echo "Cleaning up test directories..."
    rm -rf examples/output-*
    mkdir -p examples/test_suite/output
}

# Test schema generation
test_schema_generation() {
    local test_name=$1
    local input_file=$2
    local output_dir=$3
    local validator_flag=$4
    
    echo -e "\n=== Testing schema generation: $test_name ==="
    
    if [ -n "$validator_flag" ]; then
        go run main.go -input="$input_file" -output="$output_dir" -validator
    else
        go run main.go -input="$input_file" -output="$output_dir"
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Schema generation successful ✓${NC}"
        
        # Check if files were created
        if [ "$(ls -A $output_dir)" ]; then
            echo -e "${GREEN}Output files created ✓${NC}"
        else
            echo -e "${RED}No output files created ✗${NC}"
            return 1
        fi
        
        # Run KCL linting on generated files
        echo "Linting generated files..."
        find "$output_dir" -name "*.k" -exec kcl {} \; 2>&1 | grep -v "is empty json"
        if [ ${PIPESTATUS[0]} -eq 0 ]; then
            echo -e "${GREEN}Linting passed ✓${NC}"
        else
            echo -e "${RED}Linting failed ✗${NC}"
            return 1
        fi
    else
        echo -e "${RED}Schema generation failed ✗${NC}"
        return 1
    fi
    
    return 0
}

# Test validation
test_validation() {
    local test_name=$1
    local schema_file=$2
    local valid_data=$3
    local invalid_data=$4
    local schema_name=$5
    
    echo -e "\n=== Testing validation: $test_name ==="
    
    # Test valid data
    echo "Testing valid data..."
    kcl vet "$valid_data" "$schema_file" -s "$schema_name" &> /dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Valid data passed validation ✓${NC}"
    else
        echo -e "${RED}Valid data failed validation ✗${NC}"
        echo "Command: kcl vet $valid_data $schema_file -s $schema_name"
        kcl vet "$valid_data" "$schema_file" -s "$schema_name"
        return 1
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
        return 1
    fi
    
    return 0
}

# Run all tests
run_all_tests() {
    local success=true
    
    # Test 1: String constraints
    test_schema_generation "String Constraints" "examples/string_constraints.json" "examples/output-string-test" "-validator"
    [ $? -ne 0 ] && success=false
    
    # Test 2: Number constraints
    test_schema_generation "Number Constraints" "examples/number_constraints.json" "examples/output-number-test" "-validator"
    [ $? -ne 0 ] && success=false
    
    # Test 3: Array constraints
    test_schema_generation "Array Constraints" "examples/array_constraints.json" "examples/output-array-test" "-validator"
    [ $? -ne 0 ] && success=false
    
    # Test 4: Comprehensive test
    test_schema_generation "Comprehensive Test" "examples/test_suite/comprehensive_test.json" "examples/output-comprehensive-test" "-validator"
    [ $? -ne 0 ] && success=false
    
    # Test 5: Minimal test validation
    test_validation "Minimal Test" "examples/test_suite/minimal_refactored.k" \
                   "examples/test_suite/minimal_valid_data.json" \
                   "examples/test_suite/minimal_invalid_data.json" \
                   "MinimalValidator"
    [ $? -ne 0 ] && success=false
    
    # Test 6: Comprehensive test validation
    test_validation "Comprehensive Test" "examples/output-comprehensive-test/ComprehensiveTestValidator.k" \
                   "examples/test_suite/clean/valid_comprehensive_data.json" \
                   "examples/test_suite/clean/invalid_comprehensive_data.json" \
                   "ComprehensiveTestValidator"
    [ $? -ne 0 ] && success=false
    
    if $success; then
        echo -e "\n${GREEN}All tests passed successfully! ✓${NC}"
        return 0
    else
        echo -e "\n${RED}Some tests failed. ✗${NC}"
        return 1
    fi
}

# Main execution
cleanup
run_all_tests
exit_code=$?

echo -e "\n=== Test Summary ==="
if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}All tests completed successfully! ✓${NC}"
else
    echo -e "${RED}Some tests failed. Please check the output above. ✗${NC}"
fi

exit $exit_code 