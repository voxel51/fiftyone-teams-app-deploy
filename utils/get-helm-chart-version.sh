#!/usr/bin/env bash

# Looks up the version of a dependent GAR artifact from
# either the commit SHA or helm Chart.yaml version.
# There are no inputs.
# All echos happen to STDERR.
# Result is the found version which can be captured with STDOUT.

set -o pipefail
set -eu

. "$(dirname "$0")/common.sh"

version=""

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Note: Sending logs to STDERR so that users can easily use
# the STDOUT for other, more interesting automations.
if [[ ${CURRENT_BRANCH} == "main" ]]; then
  # main carries the version in flight, so take the newest released
  # <x.x.x> chart from GAR rather than the Chart.yaml version.
  log_info "Look for the newest released version in us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env" >&2
  version=$(gcloud artifacts docker images list \
    us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env \
    --include-tags \
    --format="value(tags)" | tr ';' '\n' | grep -Ex '[0-9]+\.[0-9]+\.[0-9]+' | sort -V | tail -1)
else
  # Look for the .*-<sha> in GAR.
  # Handles both x.x.x-sha-<sha> and x.x.x-rc-<sha> formats

  sha=$(git rev-parse --short HEAD)

  log_info "Look for version \"${sha}\" in us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env" >&2

  version=$(gcloud artifacts docker images list \
    us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env \
    --include-tags \
    --filter="tags ~ .*-${sha}\$" \
    --format yaml \
    --limit 1 |
    yq ".tags[0]")
fi

if [[ -z ${version} ]] || [[ ! ${version} =~ [0-9]+\.[0-9]+\.[0-9]+.* ]]; then
  log_error "No version found in the helm chart registry... Failing build.
      Please submit a PR to fiftyone-teams-app-deploy if this is a
      non-release branch so that your chart gets published.
      If this is a release branch, please view the
      'Release Pre-release charts' action in fiftyone-teams-app-deploy"
  exit 1
fi

log_info "Found internal-env chart version ${version}" >&2
# Let caller do with it what they'd like
echo "${version}"
