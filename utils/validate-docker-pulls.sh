#!/usr/bin/env bash

set -euo pipefail

VALUES_YAML=""

. "$(dirname "$0")/common.sh"

print_usage() {
  local package
  package=$(basename "$0")
  echo "$package - Validate docker pulls for all default images."
  echo " "
  echo "$package [options]"
  echo " "
  echo "options:"
  echo "-h, --help               Show brief help"
  echo "-f, --values VALUES_YAML values.yaml to use in templating"
}

parse_arguments() {
  while test $# -gt 0; do
    case "$1" in
      -h | --help)
        print_usage
        exit 0
        ;;
      -f | --values)
        if [[ -z ${2-} ]]; then
          log_error "--values requires a value"
          print_usage
          exit 1
        fi
        if [[ ! -f ${2} ]]; then
          log_error "Provided values file does not exist or does not have permissions to be read"
          print_usage
          exit 1
        fi
        VALUES_YAML="${2}"
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

cd "${GIT_ROOT}/helm/fiftyone-teams-app"

images=$(
  helm template . -f "${VALUES_YAML}" |
    yq eval -o yaml '.. | select(.image? != null) | .image' |
    sort |
    uniq |
    grep -v -- '---' |
    xargs
)

read -r -a images_array <<<"$images"

pids=()
rcs=()
exit_code=0

for img in "${images_array[@]}"; do
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
    log_error "Could not pull ${images_array[$idx]}!"
    exit_code=1
  else
    log_info "Pulled ${images_array[$idx]} successfully."
  fi
done

exit $exit_code
