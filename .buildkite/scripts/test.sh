#!/bin/bash

set -euo pipefail

echo "--- Pre install"
source .buildkite/scripts/pre-install-command.sh
go version
add_bin_path

echo "--- Go Test"
set +ex
go test -race -v ./... > tests-report.txt
exit_code=$?
set -ex

# Buildkite collapse logs under --- symbols
# need to change --- to anything else or switch off collapsing (note: not available at the moment of this commit)
echo "--- Test Results"
awk '{gsub("---", "----"); print }' tests-report.txt

exit $exit_code
