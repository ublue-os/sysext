#!/usr/bin/env bash
set -euo pipefail

export ARCH="${ARCH-x86-64}"
SCRIPTFOLDER="$(dirname "$(readlink -f "$0")")"


VERSION="$1"
SYSEXTNAME="$2"



cd "${SCRIPTFOLDER}"
"${SCRIPTFOLDER}"/bake.sh "${SYSEXTNAME}"
mkdir -p result
mv "${SYSEXTNAME}.raw" result/
rm -rf "${SYSEXTNAME}"
