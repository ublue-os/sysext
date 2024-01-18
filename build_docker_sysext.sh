#!/usr/bin/env bash
set -euo pipefail

export ARCH="${ARCH-x86-64}"
SCRIPTFOLDER="$(dirname "$(readlink -f "$0")")"
ONLY_CONTAINERD="${ONLY_CONTAINERD:-0}"
ONLY_DOCKER="${ONLY_DOCKER:-0}"

if [ $# -lt 2 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
  echo "Usage: $0 VERSION SYSEXTNAME"
  echo "The script will download the Docker release tar ball (e.g., for 20.10.13) and create a sysext squashfs image with the name SYSEXTNAME.raw in the current folder."
  echo "A temporary directory named SYSEXTNAME in the current folder will be created and deleted again."
  echo "All files in the sysext image will be owned by root."
  echo "The necessary systemd services will be created by this script, by default only docker.socket will be enabled."
  echo "To only package containerd without Docker, pass ONLY_CONTAINERD=1 as environment variable (current value is '${ONLY_CONTAINERD}')."
  echo "To only package Docker without containerd and runc, pass ONLY_DOCKER=1 as environment variable (current value is '${ONLY_DOCKER}')."
  echo "To use arm64 pass 'ARCH=arm64' as environment variable (current value is '${ARCH}')."
  "${SCRIPTFOLDER}"/bake.sh --help
  exit 1
fi

if [ "${ONLY_CONTAINERD}" = 1 ] && [ "${ONLY_DOCKER}" = 1 ]; then
  echo "Cannot set both ONLY_CONTAINERD and ONLY_DOCKER" >&2
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

rm -f "docker-${VERSION}.tgz"
curl -o "docker-${VERSION}.tgz" -fsSL "https://download.docker.com/linux/static/stable/${ARCH}/docker-${VERSION}.tgz"
# TODO: Also allow to consume upstream containerd and runc release binaries with their respective versions

tar --force-local -xf "docker-${VERSION}.tgz" -C "${SYSEXTNAME}"
rm "docker-${VERSION}.tgz"
mkdir -p "${SYSEXTNAME}"/usr/bin
mv "${SYSEXTNAME}"/docker/* "${SYSEXTNAME}"/usr/bin/
rmdir "${SYSEXTNAME}"/docker
mkdir -p "${SYSEXTNAME}/usr/lib/systemd/system"
if [ "${ONLY_CONTAINERD}" = 1 ]; then
  rm "${SYSEXTNAME}/usr/bin/docker" "${SYSEXTNAME}/usr/bin/dockerd" "${SYSEXTNAME}/usr/bin/docker-init" "${SYSEXTNAME}/usr/bin/docker-proxy"
elif [ "${ONLY_DOCKER}" = 1 ]; then
  rm "${SYSEXTNAME}/usr/bin/containerd" "${SYSEXTNAME}/usr/bin/containerd-shim-runc-v2" "${SYSEXTNAME}/usr/bin/ctr" "${SYSEXTNAME}/usr/bin/runc"
  if [[ "${VERSION%%.*}" -lt 23 ]] ; then
    # Binary releases 23 and higher don't ship containerd-shim
    rm "${SYSEXTNAME}/usr/bin/containerd-shim"
  fi
fi
rsync -av "${SCRIPTFOLDER}"/builders/docker/files/  "${SYSEXTNAME}"/

