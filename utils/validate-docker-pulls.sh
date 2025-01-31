#!/usr/bin/env bash

# Validates the pulling of docker images.
# It maintains a static list of what we expect to be pullable
# during a helm/docker deployment. It then runs a helm template,
# utilizing our default values files, to get a list of files that
# would be deployed. It first compares that list to our
# EXPECTED_IMAGES array - helping stop typos or unexpected images
# from leaking into our system. Finally, it pulls each expected image
# to make sure they are pullable.
# See ./utils/validate-docker-pulls.sh -h for help.

set -euo pipefail

# These are images we expect to be pullable in an
# exhaustive set.
EXPECTED_IMAGES=(
  "docker.io/busybox:stable-glibc" # For initContainers
  "voxel51/fiftyone-app"
  "voxel51/fiftyone-app-gpt"
  "voxel51/fiftyone-app-torch"
  "voxel51/fiftyone-teams-api"
  "voxel51/fiftyone-teams-app"
  "voxel51/fiftyone-teams-cas"
  "voxel51/fiftyone-teams-cv-full"
)

. "$(dirname "$0")/common.sh"

VALUES_YAML="${GIT_ROOT}/helm/fiftyone-teams-app/values.yaml"

print_usage() {
  local package
  package=$(basename "$0")
  echo "$package - Validate docker pulls for all default images."
  echo " "
  echo "$package [options]"
  echo " "
  echo "options:"
  echo "-h, --help               Show brief help"
  echo "-f, --values VALUES_YAML values.yaml to use in templating. Defaults to ${VALUES_YAML}"
  echo " "
  echo "examples:"
  echo "./utils/validate-docker-pulls.sh"
  echo "./utils/validate-docker-pulls.sh -f ./path/to/values.yaml"
}

parse_arguments() {
  while test $# -gt 0; do
    case "$1" in
      -h | --help)
        print_usage
        exit 0
        ;;
      -f | --values)
        if [[ -n ${2-} ]]; then
          VALUES_YAML="${2}"
        fi
        if [[ ! -f ${VALUES_YAML} ]]; then
          log_error "Provided values file does not exist or does not have permissions to be read"
          print_usage
          exit 1
        fi
        shift 2
        ;;
      *)
        log_error "Unknown option: $1"
        print_usage
        exit 1
        ;;
    esac
  done
}

check_requirements() {
  required_cmds=("helm" "yq")

  for cmd in "${required_cmds[@]}"; do
    if ! command -v "$cmd" &>/dev/null; then
      log_error "'$cmd' doesn't exist or is not executable. Please install '$cmd'"
      exit 1
    fi
  done
}

docker_pull() {
  local image="$1"

  if ! docker pull "${image}"; then
    return 1
  fi

  return 0
}

parse_arguments "$@"
check_requirements

expected_images_with_tag=()

expected_tag=$(yq '.appVersion' "${GIT_ROOT}/helm/fiftyone-teams-app/Chart.yaml")

for img in "${EXPECTED_IMAGES[@]}"; do
  if [[ ${img} =~ "voxel51/" ]]; then
    # Only add a tag for our organizations images, not
    # publicly available ones our chart may reference
    img_with_tag="${img}:${expected_tag}"
  else
    img_with_tag="${img}"
  fi
  expected_images_with_tag+=("${img_with_tag}")
done

helm_images=$(
  helm template "${GIT_ROOT}/helm/fiftyone-teams-app" -f "${VALUES_YAML}" |
    yq eval -o yaml '.. | select(.image? != null) | .image' |
    sort |
    uniq |
    grep -v -- '---' |
    xargs
)

read -r -a helm_images_array <<<"${helm_images}"

pids=()
rcs=()
exit_code=0

for img in "${helm_images_array[@]}"; do
  if [[ ! ${expected_images_with_tag[*]} =~ ${img} ]]; then
    echo "Image ${img} is in the 'helm template', but not in our published images!"
    exit 1
  fi
done

for img in "${expected_images_with_tag[@]}"; do
  docker_pull "${img}" &
  pids+=($!)
done

set +e # disable -e for 'wait'
for pid in "${pids[@]}"; do
  wait "${pid}"
  rcs+=($?)
done
set -e

for idx in "${!rcs[@]}"; do
  if [[ ${rcs[$idx]} -ne 0 ]]; then
    log_error "Could not pull ${expected_images_with_tag[$idx]}!"
    exit_code=1
  else
    log_info "Pulled ${expected_images_with_tag[$idx]} successfully."
  fi
done

exit $exit_code
