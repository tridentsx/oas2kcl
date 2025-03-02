#!/bin/bash
set -e

# Create logs directory
LOGS_DIR="test_logs"
mkdir -p "$LOGS_DIR"
LOG_FILE="$LOGS_DIR/test_run_$(date +%Y%m%d_%H%M%S).log"

# Helper function to log to both file and terminal
log() {
  echo "$@" | tee -a "$LOG_FILE"
}

# Helper for logging verbose output only to file
log_verbose() {
  echo "$@" >> "$LOG_FILE"
}

log "Running unit tests..."
go test -v ./openapikcl -run "^Test.*$" > "$LOGS_DIR/unit_tests.log" 2>&1 || {
  log "❌ Unit tests FAILED - see $LOGS_DIR/unit_tests.log for details"
  exit 1
}
log "✅ Unit tests PASSED"

# Check if the converter exists
if [ ! -f "./openapi-to-kcl" ]; then
  log "Error: openapi-to-kcl binary not found. Build it first with 'go build -o openapi-to-kcl ./cmd/main.go'"
  exit 1
fi

# Check if KCL is installed
if ! command -v kcl &> /dev/null; then
  log "Warning: KCL is not installed or not in PATH. KCL validation will be skipped."
  KCL_AVAILABLE=0
else
  log "KCL available - will perform schema validation"
  KCL_AVAILABLE=1
fi

# Create temporary directory for output
temp_dir=$(mktemp -d)
trap "rm -rf $temp_dir" EXIT

overall_result=0

log "Running integration tests..."

# Debug: List all test files
log "Available test files:"
for test_file in openapikcl/testdata/input/*.json; do
  log "  - $(basename "$test_file")"
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
    log "Error: Test file '${test_files[0]}' not found."
    exit 1
  fi
fi

log "Running tests for: ${test_files[@]}"

# For each OpenAPI file in the testdata/input directory
for input_file in "${test_files[@]}"; do
  # Skip non-existent files (in case the glob doesn't match anything)
  [ -f "$input_file" ] || continue
  
  # Extract test name from filename
  test_name=$(basename "${input_file}" .json)
  test_log="$LOGS_DIR/${test_name}_test.log"
  
  log "Testing $test_name..."
  
  # Generate KCL schemas
  ./openapi-to-kcl -oas "${input_file}" -out "${temp_dir}" -package "${test_name}" >> "$test_log" 2>&1
  
  failed=0
  
  # Run KCL validation if available
  if [ $KCL_AVAILABLE -eq 1 ]; then
    log_verbose "Running KCL validation for $test_name..."
    
    # Directly run main.k validation
    if [ -f "$temp_dir/main.k" ]; then
      log "  Running KCL validation: kcl run ./main.k"
      if (cd "$temp_dir" && kcl run "main.k" > /dev/null 2>&1); then
        log "  ✅ KCL validation: PASSED (main.k validates successfully)"
      else
        log "  ❌ KCL validation: FAILED (main.k validation errors)"
        # Show the error in the console and log
        (cd "$temp_dir" && kcl run "main.k" 2>&1) | tee -a "$test_log"
        failed=1
      fi
    else
      log "  ❌ KCL validation: FAILED (main.k file not generated)"
      failed=1
    fi
  else
    # If KCL is not available, consider the test passed if files were generated
    if [ -f "$temp_dir/main.k" ]; then
      log "  ⚠️ KCL not available - but files were generated successfully"
    else
      log "  ❌ Test FAILED (main.k file not generated)"
      failed=1
    fi
  fi
  
  if [ $failed -eq 1 ]; then
    overall_result=1
    log "❌ Test $test_name FAILED - see $test_log for details"
  else
    log "✅ Test $test_name PASSED"
  fi
done

if [ $overall_result -eq 0 ]; then
  log "All integration tests passed! 🎉"
else
  log "Some integration tests failed. See logs in $LOGS_DIR directory."
  exit 1
fi
