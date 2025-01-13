#!/usr/bin/env bash
set -euo pipefail

# Check if a variable is empty
check_empty() {
  local var_name="$1"
  local var_value="$2"
  if [[ -z $var_value ]]; then
    echo "Error: $var_name must not be empty." >&2
    print_usage
    exit 1
  fi
}
