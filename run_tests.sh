#!/bin/bash
# run_tests.sh

set -e

echo "Running unit tests..."
go test -v ./openapikcl -run "^Test.*$"

echo "All tests passed!"