#!/usr/bin/env bash

set -euo pipefail

FIFTYONE_APP_VERSION=''
FIFTYONE_TEAMS_API_VERSION=''
FIFTYONE_TEAMS_APP_VERSION=''
FIFTYONE_TEAMS_CAS_VERSION=''
DRY_RUN='false'

print_usage() {
  local package
  package=$(basename "$0")
  echo "$package - Bump versions in docker-compose fixture."
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

source ./utils/bump-fixtures-common.sh

parse_arguments() {
  while test $# -gt 0; do
    case "$1" in
      -h | --help)
        print_usage
        exit 0
        ;;
      -a | --app-version*)
        shift
        if test $# -gt 0; then
          FIFTYONE_APP_VERSION=$1
        else
          print_usage
          exit 1
        fi
        shift
        ;;
      -i | --api-version*)
        shift
        if test $# -gt 0; then
          FIFTYONE_TEAMS_API_VERSION=$1
        else
          print_usage
          exit 1
        fi
        shift
        ;;
      -t | --teams-app-version*)
        shift
        if test $# -gt 0; then
          FIFTYONE_TEAMS_APP_VERSION=$1
        else
          print_usage
          exit 1
        fi
        shift
        ;;
      -c | --cas-version*)
        shift
        if test $# -gt 0; then
          FIFTYONE_TEAMS_CAS_VERSION=$1
        else
          print_usage
          exit 1
        fi
        shift
        ;;
      -d | --dry-run)
        DRY_RUN="true"
        shift
        ;;
      *)
        break
        ;;
    esac
  done

  # Check that all version variables are set
  check_empty "FIFTYONE_APP_VERSION" "$FIFTYONE_APP_VERSION"
  check_empty "FIFTYONE_TEAMS_API_VERSION" "$FIFTYONE_TEAMS_API_VERSION"
  check_empty "FIFTYONE_TEAMS_APP_VERSION" "$FIFTYONE_TEAMS_APP_VERSION"
  check_empty "FIFTYONE_TEAMS_CAS_VERSION" "$FIFTYONE_TEAMS_CAS_VERSION"
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
  ./utils/bump-fixtures-docker.sh \
    -a "$FIFTYONE_APP_VERSION" \
    -i "$FIFTYONE_TEAMS_API_VERSION" \
    -t "$FIFTYONE_TEAMS_APP_VERSION" \
    -c "$FIFTYONE_TEAMS_CAS_VERSION" \
    -f "$fixture" $dry_run_flag
done

for fixture in "${HELM_FIXTURES[@]}"; do
  ./utils/bump-fixtures-helm.sh \
    -a "$FIFTYONE_APP_VERSION" \
    -i "$FIFTYONE_TEAMS_API_VERSION" \
    -t "$FIFTYONE_TEAMS_APP_VERSION" \
    -c "$FIFTYONE_TEAMS_CAS_VERSION" \
    -f "$fixture" $dry_run_flag
done
