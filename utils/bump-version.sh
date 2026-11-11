#!/usr/bin/env bash

# Replaces every reference to one release version with another,
# like 2.26.0 with 2.26.0,
# in the chart, Docker Compose files, skaffold.yaml, and docs.
# bump-fixtures.sh pins the integration test fixtures separately,
# so this script leaves tests/fixtures alone.

set -euo pipefail

. "$(dirname "$0")/common.sh"

FROM_VERSION=''
TO_VERSION=''

print_usage() {
  local package
  package=$(basename "$0")
  echo "$package - Replace every reference to one release version with another."
  echo " "
  echo "$package [options]"
  echo " "
  echo "options:"
  echo "-h, --help               Show brief help"
  echo "-f, --from=FROM_VERSION  Version to replace (default: version in Chart.yaml)"
  echo "-t, --to=TO_VERSION      New version, like 2.26.0 (required)"
}

parse_arguments() {
  while test $# -gt 0; do
    case "$1" in
      -h | --help)
        print_usage
        exit 0
        ;;
      -f | --from)
        if [[ -z ${2-} ]]; then
          log_error "--from requires a value"
          print_usage
          exit 1
        fi
        FROM_VERSION="${2#v}"
        shift 2
        ;;
      -t | --to)
        if [[ -z ${2-} ]]; then
          log_error "--to requires a value"
          print_usage
          exit 1
        fi
        TO_VERSION="${2#v}"
        shift 2
        ;;
      *)
        log_error "Unknown option: $1"
        print_usage
        exit 1
        ;;
    esac
  done

  if [[ -z ${TO_VERSION} ]]; then
    log_error "--to is required"
    print_usage
    exit 1
  fi

  if [[ -z ${FROM_VERSION} ]]; then
    FROM_VERSION=$(yq ".version" "${GIT_ROOT}/helm/fiftyone-teams-app/Chart.yaml")
  fi
}

parse_arguments "$@"

if [[ ${FROM_VERSION} == "${TO_VERSION}" ]]; then
  log_info "Already on ${TO_VERSION}. Nothing to change."
  exit 0
fi

cd "${GIT_ROOT}"

# git grep -F finds candidates; the perl pattern skips matches that are
# part of a longer version, like 12.25.2 or 2.25.20.
pattern="(?<![0-9.])${FROM_VERSION//./\\.}(?![0-9])"

if ! git grep --quiet -I -F "${FROM_VERSION}" -- . ':(exclude)tests/fixtures'; then
  log_error "No files reference ${FROM_VERSION}."
  exit 1
fi

git grep -z -l -I -F "${FROM_VERSION}" -- . ':(exclude)tests/fixtures' |
  xargs -0 perl -pi -e "s/${pattern}/${TO_VERSION}/g"

log_info "Replaced ${FROM_VERSION} with ${TO_VERSION} in:"
git diff --name-only
