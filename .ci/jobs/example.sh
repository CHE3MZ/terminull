#!/usr/bin/env sh
set -eu

############################################## START
printf '\033[1mSTARTING CONTAINER...\033[0m\n'
container system start 2>/dev/null || true
printf '\033[32mDONE.\033[0m\n'
############################################## START

############################################## ENV VARIABLES

export CONTAINER_IMAGE="alpine"
export CONTAINER_TAG="latest"
export CONTAINER_FULL="$CONTAINER_IMAGE:$CONTAINER_TAG"
############################################## ENV VARIABLES

############################################## JOB INT
printf '\033[34mRUNNING JOBS...\033[0m\n'
#======================================# JOB

# PULL CONTAINER IMAGE #
printf '\033[1mPULLING CONTAINER IMAGE...\033[0m\n'
container image pull "$CONTAINER_FULL"
printf '\033[32mDONE.\033[0m\n'

# ========================================= #
printf '\033[1mRUNNING CONTAINER...\033[0m\n'
# RUN CONTAINER STARTED #

container run --rm "$CONTAINER_FULL" echo "hello world"

# RUN CONTAINER STOPPED #
printf '\033[32mDONE.\033[0m\n'
# ========================================= #

# DELETE CONTAINER IMAGE #
printf '\033[1mDELETING CONTAINER IMAGE...\033[0m\n'

container image rm --force "$CONTAINER_IMAGE"
container image prune --all 2>/dev/null || true
container image prune --force --all 2>/dev/null || true
container volume prune 2>/dev/null || true
container volume prune --force -all 2>/dev/null || true
container prune 2>/dev/null || true
container container prune --force 2>/dev/null || true

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