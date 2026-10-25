#!/usr/bin/env sh
set -eu

# This script covers other automations.
# This script can be executed via task by running task other.
# Extra automations should go here instead of into the taskfile directly in order to keep the taskfile clean.

echo Running extra tasks.
#########################

sh scripts/upstream.sh
sh scripts/standalone.sh

#########################
echo Done.