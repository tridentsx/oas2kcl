#!/bin/bash
set -e

# Colors for better output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

# Create temporary directory for output
temp_dir=$(mktemp -d)
trap "rm -rf $temp_dir" EXIT

echo -e "${BLUE}===== OpenAPI to KCL Test Runner =====${RESET}"
echo -e "Using temporary directory: ${temp_dir}"

# Test results tracking
unit_tests_passed=0
integration_tests_passed=0
integration_tests_failed=0
total_schemas_validated=0
schemas_validated_success=0
schemas_validated_failed=0

echo -e "\n${BLUE}Running unit tests...${RESET}"
if go test -v ./openapikcl -run "^Test.*$"; then
  echo -e "${GREEN}✅ Unit tests PASSED${RESET}"
  unit_tests_passed=1
else
  echo -e "${RED}❌ Unit tests FAILED${RESET}"
  exit 1
fi

# Check if the converter exists
if [ ! -f "./openapi-to-kcl" ]; then
  echo -e "${RED}Error: openapi-to-kcl binary not found. Build it first with 'go build -o openapi-to-kcl ./cmd/main.go'${RESET}"
  exit 1
fi

# Check if KCL is installed
if ! command -v kcl &> /dev/null; then
  echo -e "${YELLOW}Warning: KCL is not installed or not in PATH. KCL validation will be skipped.${RESET}"
  KCL_AVAILABLE=0
else
  echo -e "${GREEN}KCL available - will perform schema validation${RESET}"
  KCL_AVAILABLE=1
fi

echo -e "\n${BLUE}Running integration tests...${RESET}"

# Debug: List all test files
echo -e "${BLUE}Available test files:${RESET}"
for test_file in openapikcl/testdata/input/*.json; do
  echo "  - $(basename "$test_file")"
done

# Use specific test files or all if "all" is specified
if [ "$1" = "all" ] || [ -z "$1" ]; then
  test_files=(
    "openapikcl/testdata/input/simple.json"
    "openapikcl/testdata/input/complex.json"
    "openapikcl/testdata/input/petstore.json"
    "openapikcl/testdata/input/complex_v2.json"
    "openapikcl/testdata/input/petstore_v2.json"
  )
else
  test_files=("openapikcl/testdata/input/$1.json")
  if [ ! -f "${test_files[0]}" ]; then
    echo -e "${RED}Error: Test file '${test_files[0]}' not found.${RESET}"
    exit 1
  fi
fi

echo -e "${BLUE}Running tests for:${RESET}"
for file in "${test_files[@]}"; do
  echo "  - $(basename "$file")"
done

# For each OpenAPI file in the testdata/input directory
for input_file in "${test_files[@]}"; do
  # Skip non-existent files (in case the glob doesn't match anything)
  [ -f "$input_file" ] || continue
  
  # Extract test name from filename
  test_name=$(basename "${input_file}" .json)
  echo -e "\n${BLUE}Testing ${test_name}...${RESET}"
  
  # Generate KCL schemas
  output_dir="${temp_dir}/${test_name}"
  mkdir -p "${output_dir}"
  
  echo "  Generating KCL schemas from ${input_file}..."
  if ./openapi-to-kcl -oas "${input_file}" -out "${output_dir}" -package "${test_name}"; then
    echo -e "  ${GREEN}✅ Schema generation successful${RESET}"
  else
    echo -e "  ${RED}❌ Schema generation failed${RESET}"
    integration_tests_failed=$((integration_tests_failed + 1))
    continue
  fi
  
  test_failed=0
  
  # Run KCL validation if available
  if [ $KCL_AVAILABLE -eq 1 ]; then
    echo -e "  ${BLUE}Running KCL validations:${RESET}"
    
    # First validate each individual schema file
    for schema_file in "${output_dir}"/*.k; do
      # Skip main.k and schema files that aren't actual KCL schemas (like helpers)
      if [[ "${schema_file}" != *"main.k"* ]] && grep -q "schema " "${schema_file}"; then
        schema_name=$(basename "${schema_file}" .k)
        total_schemas_validated=$((total_schemas_validated + 1))
        
        echo -n "    Validating ${schema_name}... "
        if (cd "${output_dir}" && kcl run "${schema_name}.k" > /dev/null 2>&1); then
          echo -e "${GREEN}PASSED${RESET}"
          schemas_validated_success=$((schemas_validated_success + 1))
        else
          echo -e "${RED}FAILED${RESET}"
          echo -e "    ${RED}Error details:${RESET}"
          (cd "${output_dir}" && kcl run "${schema_name}.k" 2>&1 | sed 's/^/      /')
          schemas_validated_failed=$((schemas_validated_failed + 1))
          test_failed=1
        fi
      fi
    done
    
    # Then validate main.k which should import and use all schemas
    if [ -f "${output_dir}/main.k" ]; then
      echo -n "    Validating main.k (all schemas)... "
      if (cd "${output_dir}" && kcl run "main.k" > /dev/null 2>&1); then
        echo -e "${GREEN}PASSED${RESET}"
      else
        echo -e "${RED}FAILED${RESET}"
        echo -e "    ${RED}Error details:${RESET}"
        (cd "${output_dir}" && kcl run "main.k" 2>&1 | sed 's/^/      /')
        test_failed=1
      fi
    else
      echo -e "    ${RED}❌ main.k file not generated${RESET}"
      test_failed=1
    fi
  else
    # If KCL is not available, consider the test passed if files were generated
    if [ -f "${output_dir}/main.k" ]; then
      echo -e "  ${YELLOW}⚠️ KCL not available - but files were generated successfully${RESET}"
    else
      echo -e "  ${RED}❌ main.k file not generated${RESET}"
      test_failed=1
    fi
  fi
  
  if [ $test_failed -eq 1 ]; then
    echo -e "  ${RED}❌ Test ${test_name} FAILED${RESET}"
    integration_tests_failed=$((integration_tests_failed + 1))
  else
    echo -e "  ${GREEN}✅ Test ${test_name} PASSED${RESET}"
    integration_tests_passed=$((integration_tests_passed + 1))
  fi
done

# Print summary
echo -e "\n${BLUE}===== Test Summary =====${RESET}"
echo -e "Unit tests: ${unit_tests_passed}/1 passed"
echo -e "Integration tests: ${integration_tests_passed}/$((integration_tests_passed + integration_tests_failed)) passed"

if [ $KCL_AVAILABLE -eq 1 ]; then
  echo -e "KCL schema validations: ${schemas_validated_success}/${total_schemas_validated} passed"
fi

if [ $integration_tests_failed -eq 0 ]; then
  echo -e "\n${GREEN}All tests passed! 🎉${RESET}"
  exit 0
else
  echo -e "\n${RED}Some tests failed.${RESET}"
  exit 1
fi
