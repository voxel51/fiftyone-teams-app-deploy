#!/usr/bin/env bash

set -eo pipefail

help() {
  echo "Get the latest image version matching the parameter value"
  echo ""
  echo "Example Usage"
  echo ""
  echo "* Get latest v1.7.0 dev version"
  echo ""
  echo "    $ /utils/get-image-versions.sh \"v1.7.0\" dev"
  echo "    fiftyone-app    v1.7.0.dev20"
  echo "    fiftyone-teams-api    v1.7.0.dev20"
  echo "    fiftyone-teams-app    v1.7.0-dev.16"
  echo "    fiftyone-teams-cas    v1.7.0-dev.1"
  echo ""
  echo " * Get latest v1.7.0 rc version"
  echo "    $ ./utils/get-image-versions.sh \"v1.7.0\" rc"
  echo "    fiftyone-app    v1.7.0rc8"
  echo "    fiftyone-teams-api    v1.7.0rc8"
  echo "    fiftyone-teams-app    v1.7.0-rc.7"
  echo "    fiftyone-teams-cas    v1.7.0-rc.7"
  echo ""
  echo "* Get latest versions"
  echo ""
  echo "    $ ./utils/get-image-versions.sh latest"
  echo "    fiftyone-app    v1.8.0.dev14"
  echo "    fiftyone-teams-api    v1.8.0.dev14"
  echo "    fiftyone-teams-app    v1.8.0-dev.14"
  echo "    fiftyone-teams-cas    v1.8.0-dev.14"
}

if [ "${1}" == 'help' ]; then
  help
  exit 0
fi

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

  _VERSION_REGEX="${2}"
  _FILTER_KEY="${3}"

  # Get latest image version matching the regex pattern
  _VERSION=$(
    gcloud artifacts tags list \
      --location=us-central1 \
      --repository=dev-docker \
      --project computer-vision-team \
      --package "${PACKAGE}" \
      --filter="${_FILTER_KEY} ~ ${_VERSION_REGEX}" \
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
  if [ "${VERSION}" == "latest" ]; then
    # Get last package, sorted descending by creation time
    _SHA=$(
      gcloud artifacts versions list \
        --location=us-central1 \
        --repository=dev-docker \
        --project computer-vision-team \
        --package "${IMAGE}" \
        --sort-by ~CREATE_TIME \
        --limit 1 \
        --format="csv[no-heading](name)" \
        2>/dev/null
    )
    get_latest_image "${IMAGE}" "${_SHA}" version
  else
    get_latest_image "${IMAGE}" "${VERSION_REGEX}" name
  fi
done
