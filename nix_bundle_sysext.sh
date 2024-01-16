#!/usr/bin/env bash
set -euo pipefail

export ARCH="${ARCH-x86-64}"
SCRIPTFOLDER="$(dirname "$(readlink -f "$0")")"

if [ $# -lt 2 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
  echo "Usage: $0  SYSEXTNAME"
  echo "The script will export a nix derivation for FLAKEREF and create a sysext squashfs image with the name SYSEXTNAME.raw in the current folder."
  echo "A temporary directory named SYSEXTNAME in the current folder will be created and deleted again."
  echo "All files in the sysext image will be owned by root."
  echo "To use arm64 pass 'ARCH=arm64' as environment variable (current value is '${ARCH}')."
  "${SCRIPTFOLDER}"/bake.sh --help
  exit 1
fi

FLAKEREF="$1"
SYSEXTNAME="$2"

# The github release uses different arch identifiers, we map them here
# and rely on bake.sh to map them back to what systemd expects
if [ "${ARCH}" = "amd64" ] || [ "${ARCH}" = "x86-64" ]; then
  ARCH="x86_64"
elif [ "${ARCH}" = "arm64" ]; then
  ARCH="aarch64"
fi

# clean target
rm -rf "${SYSEXTNAME}"
mkdir -p "${SYSEXTNAME}"

cd "${SYSEXTNAME}"
#nix things here
nix bundle "${FLAKEREF}" --extra-experimental-features nix-command --extra-experimental-features flakes
pwd
ls -la
target=$(readlink -f "${SYSEXTNAME}")
echo $target

cd "${SCRIPTFOLDER}"
mkdir -p "${SYSEXTNAME}"/usr/local/bin
cp "${target}" "${SYSEXTNAME}"/usr/local/bin/"${SYSEXTNAME}"

cd "${SCRIPTFOLDER}"

"${SCRIPTFOLDER}"/bake.sh "${SYSEXTNAME}"
mkdir -p result
mv "${SYSEXTNAME}.raw" result/
rm -rf "${SYSEXTNAME}"
