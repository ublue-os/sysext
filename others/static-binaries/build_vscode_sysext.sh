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

VERSION="$1"
SYSEXTNAME="$2"

# The github release uses different arch identifiers, we map them here
# and rely on sysext.sh to map them back to what systemd expects
if [ "${ARCH}" = "amd64" ] || [ "${ARCH}" = "x86-64" ]; then
  ARCH="x86_64"
elif [ "${ARCH}" = "arm64" ]; then
  ARCH="aarch64"
fi

# clean target


cd "${SYSEXTNAME}"
curl -o "vscode.tar.gz" -fsSL "https://code.visualstudio.com/sha/download?build=stable&os=linux-x64"

tar -xzf vscode.tar.gz
rm vscode.tar.gz
mkdir -p "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin/share/
mv VSCode-linux-x64 "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin/share/code
mkdir -p "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin/bin
mkdir -p "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin/share/icons
cd "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin/bin
ln -s ../share/code/bin/code code

cd "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin/share/icons
ln -s  ../code/resources/app/resources/linux/code.png vscode.png
cd "${SCRIPTFOLDER}/${SYSEXTNAME}"

