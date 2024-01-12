#!/usr/bin/env bash
set -euo pipefail

export ARCH="${ARCH-x86-64}"
SCRIPTFOLDER="$(dirname "$(readlink -f "$0")")"

if [ $# -lt 2 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
  echo "Usage: $0 VERSION SYSEXTNAME"
  echo "The script will download the wasmtime release tar ball (e.g., for 4.0.0) and create a sysext squashfs image with the name SYSEXTNAME.raw in the current folder."
  echo "A temporary directory named SYSEXTNAME in the current folder will be created and deleted again."
  echo "All files in the sysext image will be owned by root."
  echo "To use arm64 pass 'ARCH=arm64' as environment variable (current value is '${ARCH}')."
  "${SCRIPTFOLDER}"/bake.sh --help
  exit 1
fi

VERSION="$1"
SYSEXTNAME="$2"

# The github release uses different arch identifiers, we map them here
# and rely on bake.sh to map them back to what systemd expects
if [ "${ARCH}" = "amd64" ] || [ "${ARCH}" = "x86-64" ]; then
  ARCH="x86_64"
elif [ "${ARCH}" = "arm64" ]; then
  ARCH="aarch64"
fi

rm -f "go${VERSION}.linux-amd64.tar.xz"

curl -o "go${VERSION}.linux-amd64.tar.xz" -fsSL "https://go.dev/dl/go${VERSION}.linux-amd64.tar.gz"
rm -rf "${SYSEXTNAME}"
mkdir -p "${SYSEXTNAME}"/usr/local
tar --force-local -xzf "go${VERSION}.linux-amd64.tar.xz" -C "${SYSEXTNAME}"/usr/local
rm -f "go${VERSION}.linux-amd64.tar.xz"
"${SCRIPTFOLDER}"/bake.sh "${SYSEXTNAME}"
mkdir -p result
mv "${SYSEXTNAME}.raw" result/
rm -rf "${SYSEXTNAME}"