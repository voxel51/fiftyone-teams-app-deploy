#!/usr/bin/env bash

RED="\e[31m"
GREEN="\e[32m"
NO_COLOR="\e[0m"

# shellcheck disable=SC2034
GIT_ROOT=$(git rev-parse --show-toplevel)

log_info() {
  local msg="${1}"
  printf "%sINFO: %s%s$\n" "${GREEN}" "${msg}" "${NO_COLOR}"
}

log_error() {
  local msg="${1}"
  printf "%sERROR: %s%s\n" "${RED}" "${msg}" "${NO_COLOR}" >&2
}
