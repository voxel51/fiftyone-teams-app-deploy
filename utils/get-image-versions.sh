#!/usr/bin/env bash

# Get the latest image version matching the parameter value

set -eo pipefail

IMAGES=(
  fiftyone-app
  fiftyone-teams-api
  fiftyone-teams-app
  fiftyone-teams-cas
)

if [ -z "${1}" ]; then
  echo "ERROR - Provide package version regex. Ex: 1.7.0"
else
  VERSION="${1}"
fi

SEGMENT="${2}"

get_latest_image() {
  # Validate pameter
  if [ -z "${1}" ]; then
    echo "ERROR - Provide package name"
    exit 10
  else
    PACKAGE="${1}"
  fi

  _REGEX_VERSION=${2}

  # Get latest image version matching the regex pattern
  _VERSION=$(
    gcloud artifacts tags list \
      --location=us-central1 \
      --repository=dev-docker \
      --project computer-vision-team \
      --package "${PACKAGE}" \
      --filter="name ~ ${_REGEX_VERSION}" \
      2>/dev/null |
      sort -V |
      tail -n 1 |
      cut -f 1 -d ' '
  )

  if [ -z "${_VERSION}" ]; then
    _VERSION="Not found"
  fi

  # Display package and version
  printf "%s\t%s\n" "${1}" "${_VERSION}"
}

# Construct regex for version accounting for the different tagging conventions
case "${SEGMENT}" in
  rc)
    VERSION_REGEX="${VERSION}*.${SEGMENT}*"
    ;;

  dev)
    VERSION_REGEX="${VERSION}*.${SEGMENT}*"
    ;;

  beta)
    VERSION_REGEX="${VERSION}*.${SEGMENT}*"
    ;;

  "")
    VERSION_REGEX="${VERSION}$"
    ;;

  *)
    VERSION_REGEX="${VERSION}$"
    ;;
esac

# Call function with package and regex pattern
for IMAGE in "${IMAGES[@]}"; do
  get_latest_image "${IMAGE}" "${VERSION_REGEX}"
done
