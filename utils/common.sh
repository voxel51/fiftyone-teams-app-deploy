#!/usr/bin/env bash

RED="\033[0;31m"
GREEN="\033[0;32m"
NO_COLOR="\033[0m"

# shellcheck disable=SC2034
GIT_ROOT=$(git rev-parse --show-toplevel)

log_info() {
  local msg="${1}"
  echo -e "${GREEN}INFO: ${msg}${NO_COLOR}"
}

log_error() {
  local msg="${1}"
  echo -e "${RED}ERROR: ${msg}${NO_COLOR}" >&2
}
