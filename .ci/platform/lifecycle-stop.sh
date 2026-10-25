#!/usr/bin/env sh
set -eu

############################################## ENV VARIABLES
export CONTAINER_IMAGE="alpine"
export CONTAINER_TAG="latest"
export CONTAINER_FULL="$CONTAINER_IMAGE:$CONTAINER_TAG"
export OS=$(fastfetch -s OS --format json | jq -r '.[0].result.name' | tr '[:upper:]' '[:lower:]')
############################################## ENV VARIABLES

# RUN CONTAINER STOPPED #
printf '\033[32mDONE.\033[0m\n'
# ========================================= #

# DELETE CONTAINER IMAGE #
printf '\033[1mDELETING CONTAINER IMAGE...\033[0m\n'

container image rm --force "$CONTAINER_IMAGE"

if [ "$OS" = "windows" ]; then
    echo detected "$OS"...
    container image prune --force --all
    container volume prune --force --all
    container container prune --force
elif [ "$OS" = "macos" ]; then
    echo detected "$OS"...
    container image prune --all
    container volume prune
    container prune
else
  echo "unknown platform"
fi

printf '\033[32mDONE.\033[0m\n'

#======================================# JOB
printf '\033[34mALL JOBS WERE RAN.\033[0m\n'
############################################## JOB END

############################################## STOP
printf '\033[1mSTOPPING CONTAINER...\033[0m\n'
container system stop 2>/dev/null || true
container system session terminate 2>/dev/null || true
printf '\033[32mDONE.\033[0m\n'
############################################## STOP