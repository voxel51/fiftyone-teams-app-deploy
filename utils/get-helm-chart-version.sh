#!/usr/bin/env bash

# Looks up the version of a dependent GAR artifact from
# either the commit SHA or helm Chart.yaml version.
# There are no inputs.
# All echos happen to STDERR.
# Result is the found version which can be captured with STDOUT.

set -o pipefail
set -eu

version=""
chart_ver=$(yq ".version" helm/fiftyone-teams-app/Chart.yaml)

# Note: Sending echos to STDERR so that users can easily use
# the STDOUT for other, more interesting automations.
if [[ ${TEAMS_DEPLOYER_BRANCH} == "main" ]]; then
  # Get the most recent <x.x.x> version from GAR
  pattern="${chart_ver}"
  echo "Look for version \"${pattern}\" in us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env" >&2
  version=$(gcloud artifacts docker images list \
    us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env \
    --include-tags \
    --format="value(tags)" \
    --filter="tags:${pattern}" \
    --sort-by createTime | grep -E '[0-9]+\.[0-9]+\.[0-9]+$$' | tail -1)
else
  # Look for the .*-<sha> in GAR.
  # Handles both x.x.x-sha-<sha> and x.x.x-rc-<sha> formats

  sha=$(git rev-parse --short HEAD)

  echo "Look for version \"${sha}\" in us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env" >&2

  version=$(gcloud artifacts docker images list \
    us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env \
    --include-tags \
    --filter="tags ~ .*-${sha}\$" \
    --format yaml \
    --limit 1 |
    yq ".tags[0]")
fi

if [[ -z ${version} ]] || [[ ! ${version} =~ [0-9]+\.[0-9]+\.[0-9]+.* ]]; then
  echo "[ERROR] No version found in the helm chart registry... Failing build." >&2
  echo "        Please submit a PR to fiftyone-teams-app-deploy if this is a " >&2
  echo "        non-release branch so that your chart gets published." >&2
  echo "        If this is a release branch, please view the " >&2
  echo "        'Release Pre-release charts' action in fiftyone-teams-app-deploy" >&2
  exit 1
fi

echo "Found internal-env chart version ${version}" >&2
# Let caller do with it what they'd like
echo "${version}"
