#!/usr/bin/env bash
set -euo pipefail

export ARCH="${ARCH-x86-64}"
SCRIPTFOLDER="$(dirname "$(readlink -f "$0")")"

if [ $# -lt 2 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
  echo "Usage: $0 VERSION SYSEXTNAME"
  echo "The script will download the Incus release tar ball (e.g., for 0.4) and create a sysext squashfs image with the name SYSEXTNAME.raw in the current folder."
  echo "A temporary directory named SYSEXTNAME in the current folder will be created and deleted again."
  echo "All files in the sysext image will be owned by root."
  echo "The necessary systemd services will be created by this script, by default only docker.socket will be enabled."
  echo "To use arm64 pass 'ARCH=arm64' as environment variable (current value is '${ARCH}')."
  "${SCRIPTFOLDER}"/sysext.sh --help
  exit 1
fi

git config --global --add safe.directory /bakery  

VERSION="$1"
SYSEXTNAME="$2"

pwd
ls -la
. ./builders/"${SYSEXTNAME}"/env.sh

# The github release uses different arch identifiers, we map them here
# and rely on sysext.sh to map them back to what systemd expects
if [ "${ARCH}" = "amd64" ] || [ "${ARCH}" = "x86-64" ]; then
  ARCH="x86_64"
elif [ "${ARCH}" = "arm64" ]; then
  ARCH="aarch64"
fi

# clean target
rm -rf "${SYSEXTNAME}"
mkdir -p "${SYSEXTNAME}"

BLUEFINPREFIX="/usr/bluefin"
"${SCRIPTFOLDER}"/build_vscode_sysext.sh "${VERSION}" "${SYSEXTNAME}"
"${SCRIPTFOLDER}"/build_neovim_sysext.sh 0.9.5 "${SYSEXTNAME}"
"${SCRIPTFOLDER}"/build_docker_sysext.sh 24.0.6 "${SYSEXTNAME}"
rsync -av "${SCRIPTFOLDER}"/builders/meta/files/  "${SYSEXTNAME}"/


cd "${SCRIPTFOLDER}"
"${SCRIPTFOLDER}"/finalize.sh "${VERSION}" "${SYSEXTNAME}"

