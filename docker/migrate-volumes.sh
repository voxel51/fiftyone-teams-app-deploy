#!/usr/bin/env bash
# Copy the plugins volume from an older Docker Compose project into this one.
#
# Before v2.26, deployments ran from docker/internal-auth or docker/legacy-auth,
# so Compose named the plugins volume after that directory
# (for example, internal-auth_plugins-vol).
# docker/compose.yaml now names the project voxel51,
# so Compose would start with a new, empty plugins volume.
# This script finds the old volume by its Compose labels
# and copies its contents into the new project's volume.
# It picks a source on its own only from the internal-auth or legacy-auth projects,
# because a host can run other deployments with their own plugins volumes.
# It never changes or removes the old volume.

set -euo pipefail

VOLUME_KEY='plugins-vol'
# Project names that the pre-v2.26 docker/internal-auth and
# docker/legacy-auth directories gave their deployments by default.
DEFAULT_OLD_PROJECTS=(internal-auth legacy-auth)
FROM_PROJECT=''
IMAGE=''
FORCE='false'
DRY_RUN='false'
COMPOSE_ARGS=()

print_usage() {
  local package
  package=$(basename "$0")
  cat <<EOF
$package - Copy the plugins volume from an older Compose project into this one.

Run it from the docker directory after you stop the old stack
and before you start the new one.
Pass the same docker compose options you use for 'up' after '--'.

$package [options] -- [docker compose options]

options:
  --from-project NAME  copy from this Compose project's plugins volume
                       (required unless the old project is internal-auth
                       or legacy-auth)
  --image IMAGE        image used to copy the files
                       (default: the telemetry-redis service's image)
  --force              copy even when the new plugins volume already has files
  --dry-run            print what would happen without changing anything
  -h, --help           show this help

example:
  $package -- \\
    -f compose.yaml \\
    -f compose.plugins.yaml \\
    -f compose.override.yaml
EOF
}

log() {
  echo "$package: $*"
}

fail() {
  echo "$package: error: $*" >&2
  exit 1
}

package=$(basename "$0")

while [[ $# -gt 0 ]]; do
  case $1 in
    --from-project)
      [[ $# -ge 2 ]] || fail "--from-project needs a project name"
      FROM_PROJECT=$2
      shift 2
      ;;
    --image)
      [[ $# -ge 2 ]] || fail "--image needs an image name"
      IMAGE=$2
      shift 2
      ;;
    --force)
      FORCE='true'
      shift
      ;;
    --dry-run)
      DRY_RUN='true'
      shift
      ;;
    -h | --help)
      print_usage
      exit 0
      ;;
    --)
      shift
      COMPOSE_ARGS=("$@")
      break
      ;;
    *)
      print_usage >&2
      fail "unknown option: $1"
      ;;
  esac
done

compose() {
  docker compose ${COMPOSE_ARGS[@]+"${COMPOSE_ARGS[@]}"} "$@"
}

# Print the label value of a volume.
volume_label() {
  docker volume inspect \
    --format "{{ index .Labels \"$2\" }}" \
    "$1"
}

# Print the image of the telemetry-redis service from the rendered config.
telemetry_redis_image() {
  awk '
    /^  telemetry-redis:$/ { in_service = 1; next }
    in_service && /^  [^ ]/ { exit }
    in_service && /^    image: / { print $2; exit }
  ' <<<"$1"
}

run_or_print() {
  if [[ $DRY_RUN == 'true' ]]; then
    log "dry run: $*"
  else
    "$@"
  fi
}

config=$(compose config) ||
  fail "'docker compose config' failed. Check the options after '--'."

project=$(sed -n 's/^name: //p' <<<"$config" | head -n 1)
[[ -n $project ]] || fail "could not read the project name from 'docker compose config'"

volumes=$(compose config --volumes) ||
  fail "'docker compose config --volumes' failed. Check the options after '--'."
if ! grep -qx "$VOLUME_KEY" <<<"$volumes"; then
  log "skipped: these compose files don't use $VOLUME_KEY. Add compose.plugins.yaml or compose.dedicated-plugins.yaml if you use plugins."
  exit 0
fi

dest="${project}_${VOLUME_KEY}"

image_checked='false'
ensure_image() {
  [[ $image_checked == 'false' ]] || return 0
  if [[ -z $IMAGE ]]; then
    IMAGE=$(telemetry_redis_image "$config")
    [[ -n $IMAGE ]] || fail "could not find the telemetry-redis image. Pass --image."
  fi
  if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
    fail "image '$IMAGE' isn't on this host. Run 'docker compose <your options> pull' first, load it on air-gapped hosts, or pass --image."
  fi
  image_checked='true'
}

is_default_old_project() {
  local name
  for name in "${DEFAULT_OLD_PROJECTS[@]}"; do
    [[ $1 == "$name" ]] && return 0
  done
  return 1
}

# A project that kept its name, or a second run, already has its plugins.
dest_exists='false'
if docker volume inspect "$dest" >/dev/null 2>&1; then
  dest_exists='true'
  ensure_image
  contents=$(docker run \
    --rm \
    --volume "$dest:/to:ro" \
    --entrypoint sh \
    "$IMAGE" \
    -c 'ls -A /to | head -n 1')
  if [[ -n $contents && $FORCE != 'true' ]]; then
    log "skipped: $dest already has files, so there's nothing to copy. Use --force to copy over them."
    exit 0
  fi
fi

# Find plugins volumes that other Compose projects created.
plugins_volumes=$(docker volume ls --quiet --filter "label=com.docker.compose.volume=$VOLUME_KEY") ||
  fail "'docker volume ls' failed. Check that this user can run docker commands."
sources=()
source_projects=()
other_projects=()
while IFS= read -r volume; do
  [[ -n $volume ]] || continue
  volume_project=$(volume_label "$volume" com.docker.compose.project)
  [[ $volume_project != "$project" ]] || continue
  if [[ -n $FROM_PROJECT ]]; then
    [[ $volume_project == "$FROM_PROJECT" ]] || continue
  elif ! is_default_old_project "$volume_project"; then
    other_projects+=("$volume_project")
    continue
  fi
  sources+=("$volume")
  source_projects+=("$volume_project")
done <<<"$plugins_volumes"

if [[ ${#sources[@]} -eq 0 ]]; then
  if [[ -n $FROM_PROJECT ]]; then
    fail "no $VOLUME_KEY volume belongs to project '$FROM_PROJECT'"
  fi
  log "skipped: no $VOLUME_KEY volume from an internal-auth or legacy-auth project was found"
  if [[ ${#other_projects[@]} -gt 0 ]]; then
    log "other projects on this host have one: ${other_projects[*]}
  If one of them is this deployment's old project, run again with --from-project <project>."
  fi
  exit 0
fi

if [[ ${#sources[@]} -gt 1 ]]; then
  echo "$package: error: more than one older project has a $VOLUME_KEY volume:" >&2
  for i in "${!sources[@]}"; do
    echo "  project ${source_projects[$i]}: volume ${sources[$i]}" >&2
  done
  fail "choose one with --from-project <project>"
fi

src=${sources[0]}
src_project=${source_projects[0]}
log "found $src from project '$src_project'"

running=$(docker ps --quiet --filter "label=com.docker.compose.project=$src_project")
if [[ -n $running ]]; then
  fail "project '$src_project' is still running. Stop it first, without -v so its volumes stay:
  docker compose -p $src_project down"
fi

ensure_image

if [[ $dest_exists == 'false' ]]; then
  log "creating $dest"
  # Compose reuses a volume without a warning when these two labels match.
  run_or_print docker volume create \
    --label "com.docker.compose.project=$project" \
    --label "com.docker.compose.volume=$VOLUME_KEY" \
    "$dest"
fi

log "copying $src to $dest"
run_or_print docker run \
  --rm \
  --volume "$src:/from:ro" \
  --volume "$dest:/to" \
  --entrypoint sh \
  "$IMAGE" \
  -c 'cp -a /from/. /to/'

if [[ $DRY_RUN == 'true' ]]; then
  log "dry run: nothing changed"
  exit 0
fi

log "done. Once the new stack works, you can remove the old volume:
  docker volume rm $src"
