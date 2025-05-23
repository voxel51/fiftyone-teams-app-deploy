#!/usr/bin/env bash

set -euo pipefail

FIFTYONE_APP_VERSION=''
FIFTYONE_TEAMS_API_VERSION=''
FIFTYONE_TEAMS_APP_VERSION=''
FIFTYONE_TEAMS_CAS_VERSION=''
DRY_RUN='false'

GIT_ROOT=$(git rev-parse --show-toplevel)

print_usage() {
  local package
  package=$(basename "$0")
  echo "$package - Bump versions in Docker Compose and helm fixtures."
  echo " "
  echo "$package [options]"
  echo " "
  echo "options:"
  echo "-h, --help                                          show brief help"
  echo "-a, --app-version=FIFTYONE_APP_VERSION              Set Fiftyone App Version"
  echo "-i, --api-version=FIFTYONE_TEAMS_API_VERSION        Set Fiftyone Teams API Version"
  echo "-t, --teams-app-version=FIFTYONE_TEAMS_APP_VERSION  Set Fiftyone Teams App Version"
  echo "-c, --cas-version=FIFTYONE_TEAMS_CAS_VERSION        Set Fiftyone CAS Version"
  echo "-d, --dry-run                                       Perform a dry-run (print to stdout instead of modifying the file)"
}

source "${GIT_ROOT}/utils/bump-fixtures-common.sh"

parse_arguments() {
  while test $# -gt 0; do
    case "$1" in
      -h | --help)
        print_usage
        exit 0
        ;;
      -a | --app-version)
        if [[ -z ${2-} ]]; then
          echo "Error: --app-version requires a value" >&2
          print_usage
          exit 1
        fi
        FIFTYONE_APP_VERSION="${2}"
        shift 2
        ;;
      -i | --api-version)
        if [[ -z ${2-} ]]; then
          echo "Error: --api-version requires a value" >&2
          print_usage
          exit 1
        fi
        FIFTYONE_TEAMS_API_VERSION="${2}"
        shift 2
        ;;
      -t | --teams-app-version)
        if [[ -z ${2-} ]]; then
          echo "Error: --teams-app-version requires a value" >&2
          print_usage
          exit 1
        fi
        FIFTYONE_TEAMS_APP_VERSION="${2}"
        shift 2
        ;;
      -c | --cas-version)
        if [[ -z ${2-} ]]; then
          echo "Error: --cas-version requires a value" >&2
          print_usage
          exit 1
        fi
        FIFTYONE_TEAMS_CAS_VERSION="${2}"
        shift 2
        ;;
      -d | --dry-run)
        DRY_RUN="true"
        shift
        ;;
      *)
        echo "Error: Unknown option: $1" >&2
        print_usage
        exit 1
        ;;
    esac
  done

  # Check that all version variables are set
  check_empty "FIFTYONE_APP_VERSION" "${FIFTYONE_APP_VERSION}"
  check_empty "FIFTYONE_TEAMS_API_VERSION" "${FIFTYONE_TEAMS_API_VERSION}"
  check_empty "FIFTYONE_TEAMS_APP_VERSION" "${FIFTYONE_TEAMS_APP_VERSION}"
  check_empty "FIFTYONE_TEAMS_CAS_VERSION" "${FIFTYONE_TEAMS_CAS_VERSION}"
}

# Parse the arguments
parse_arguments "$@"

DOCKER_FIXTURES=(
  tests/fixtures/docker/integration_legacy_auth.env
  tests/fixtures/docker/integration_internal_auth.env
)

HELM_FIXTURES=(
  tests/fixtures/helm/integration_values.yaml
)

dry_run_flag=""
if [[ $DRY_RUN == "true" ]]; then
  dry_run_flag="-d"
fi

for fixture in "${DOCKER_FIXTURES[@]}"; do
  "${GIT_ROOT}/utils/bump-fixtures-docker.sh" \
    -a "${FIFTYONE_APP_VERSION}" \
    -i "${FIFTYONE_TEAMS_API_VERSION}" \
    -t "${FIFTYONE_TEAMS_APP_VERSION}" \
    -c "${FIFTYONE_TEAMS_CAS_VERSION}" \
    -f "${fixture}" $dry_run_flag
done

for fixture in "${HELM_FIXTURES[@]}"; do
  "${GIT_ROOT}/utils/bump-fixtures-helm.sh" \
    -a "${FIFTYONE_APP_VERSION}" \
    -i "${FIFTYONE_TEAMS_API_VERSION}" \
    -t "${FIFTYONE_TEAMS_APP_VERSION}" \
    -c "${FIFTYONE_TEAMS_CAS_VERSION}" \
    -f "${fixture}" $dry_run_flag
done
