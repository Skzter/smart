#!/bin/bash

set -e

THRESHOLD=$1
COVERAGE_FILE=$2

if [ -z "$THRESHOLD" ]; then
  echo "Error: Threshold not provided."
  exit 1
fi

if [ -z "$COVERAGE_FILE" ]; then
    echo "Error: Coverage file not provided."
    exit 1
fi

totalCoverage=$(go tool cover -func="$COVERAGE_FILE" | grep total | grep -Eo '[0-9]+\.[0-9]+')

echo "Current test coverage: $totalCoverage %"
if (( $(echo "$totalCoverage $THRESHOLD" | awk '{print ($1 > $2)}') )); then
    echo "Coverage check passed"
else
    echo "Current test coverage is below threshold."
    exit 1
fi 