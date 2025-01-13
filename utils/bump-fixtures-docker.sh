#!/usr/bin/env bash

set -euo pipefail

# Default values for the version variables (can be empty, but will need to be set by the user)
FIFTYONE_APP_VERSION=''
FIFTYONE_TEAMS_API_VERSION=''
FIFTYONE_TEAMS_APP_VERSION=''
FIFTYONE_TEAMS_CAS_VERSION=''
INPUT_FILE=''
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
  echo "-f, --file=INPUT_FILE                               .env file to update"
  echo "-d, --dry-run                                       Perform a dry-run (print to stdout instead of modifying the file)"
}

# Parse command-line options
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
      -f | --file*)
        shift
        INPUT_FILE="$1"
        if [[ ! -f $INPUT_FILE ]]; then
          echo "Error: File '$INPUT_FILE' does not exist." >&2
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
  check_empty "INPUT_FILE" "$INPUT_FILE"
}

source ./utils/bump-fixtures-common.sh

# Parse the arguments
parse_arguments "$@"

# Set up temporary file handling for dry run
file="$INPUT_FILE"
if [[ $DRY_RUN == "true" ]]; then
  tempfile=$(mktemp)
  cp "$INPUT_FILE" "$tempfile"
  file="$tempfile"
  echo "Performing dry-run: Changes will be printed but not saved."
fi

# Determine the appropriate `sed` flags based on the OS type
sed_flags="-i"
delete_backups="false"
if [[ $OSTYPE == "darwin"* ]]; then
  sed_flags="-ib" # macOS requires a backup extension when using `-i`
  delete_backups="true"
fi

# Perform replacements in the file
sed "$sed_flags" "s/^FIFTYONE_APP_VERSION=.*/FIFTYONE_APP_VERSION=$FIFTYONE_APP_VERSION/" "$file"
sed "$sed_flags" "s/^FIFTYONE_TEAMS_API_VERSION=.*/FIFTYONE_TEAMS_API_VERSION=$FIFTYONE_TEAMS_API_VERSION/" "$file"
sed "$sed_flags" "s/^FIFTYONE_TEAMS_APP_VERSION=.*/FIFTYONE_TEAMS_APP_VERSION=$FIFTYONE_TEAMS_APP_VERSION/" "$file"
sed "$sed_flags" "s/^FIFTYONE_TEAMS_CAS_VERSION=.*/FIFTYONE_TEAMS_CAS_VERSION=$FIFTYONE_TEAMS_CAS_VERSION/" "$file"

# Output the file contents (dry-run will print the content)
cat "$file"

# Clean up backup file if on macOS
if [[ $delete_backups == "true" ]]; then
  rm "${file}b"
fi

# Remove temporary file if dry-run
if [[ $DRY_RUN == "true" ]]; then
  rm "$tempfile"
fi
