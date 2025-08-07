#!/usr/bin/env bash

set -o pipefail
set -eu

version=""
chart_ver=$(yq ".version" helm/fiftyone-teams-app/Chart.yaml)

if [[ ${TEAMS_DEPLOYER_BRANCH} == "main" ]]; then
  # Get the most recent <x.x.x> version from GAR
  pattern="${chart_ver}"
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

  echo "Look for version \"${sha}\" in us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env"

  version=$(gcloud artifacts docker images list \
    us-central1-docker.pkg.dev/computer-vision-team/helm-internal/internal-env \
    --include-tags \
    --filter="tags ~ .*-${sha}\$" \
    --format yaml \
    --limit 1 |
    yq ".tags[0]")
fi

if [[ -z ${version} ]] || [[ ! ${version} =~ [0-9]+\.[0-9]+\.[0-9]+.* ]]; then
  echo "[ERROR] No version found in the helm chart registry... Failing build."
  echo "        Please submit a PR to fiftyone-teams-app-deploy if this is a "
  echo "        non-release branch so that your chart gets published."
  echo "        If this is a release branch, please view the "
  echo "        'Release Pre-release charts' action in fiftyone-teams-app-deploy"
  exit 1
fi

# Let caller do with it what they'd like
echo "${version}"
