#!/usr/bin/env bash
set -euo pipefail

export ARCH="${ARCH-x86-64}"
SCRIPTFOLDER="$(dirname "$(readlink -f "$0")")"


VERSION="$1"
SYSEXTNAME="$2"



cd "${SCRIPTFOLDER}"
# make the sysext
"${SCRIPTFOLDER}"/sysext.sh "${SYSEXTNAME}"

# remove the usr folder so only the etc folder is left
rm -rf "${SYSEXTNAME}"/usr

ls -la "${SYSEXTNAME}"
# make the result folder
mkdir -p result

# if there are files in SYSEXTNAME/etc, call the confext.sh script too
if [ "$(ls -A "${SYSEXTNAME}"/etc)" ]; then
  "${SCRIPTFOLDER}"/confext.sh "${SYSEXTNAME}"
  # move the result to the result folder
  mv "${SYSEXTNAME}.confext.raw" result/
fi

mv "${SYSEXTNAME}.sysext.raw" result/
rm -rf "${SYSEXTNAME}"
