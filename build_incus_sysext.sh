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
  "${SCRIPTFOLDER}"/bake.sh --help
  exit 1
fi

git config --global --add safe.directory /bakery  

VERSION="$1"
SYSEXTNAME="$2"

pwd
ls -la
. ./builders/"${SYSEXTNAME}"/env.sh

# The github release uses different arch identifiers, we map them here
# and rely on bake.sh to map them back to what systemd expects
if [ "${ARCH}" = "amd64" ] || [ "${ARCH}" = "x86-64" ]; then
  ARCH="x86_64"
elif [ "${ARCH}" = "arm64" ]; then
  ARCH="aarch64"
fi



BLUEFINPREFIX="/usr/bluefin"

cd "${SYSEXTNAME}"
git clone https://github.com/cowsql/raft.git
cd raft
git checkout "v${RAFT_VERSION}"
autoreconf -i
./configure --prefix="${BLUEFINPREFIX}"
make
make install DESTDIR="${SCRIPTFOLDER}/${SYSEXTNAME}"
ls -la "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin


cd "${SCRIPTFOLDER}"
pwd
export GOPATH="${SCRIPTFOLDER}/${SYSEXTNAME}"
mkdir -p "${SCRIPTFOLDER}/${SYSEXTNAME}"/usr/bluefin/include
# cowsql
rm -f "cowsql-${COWSQL_VERSION}.tar.gz"
curl -o "cowsql-${COWSQL_VERSION}.tar.gz" -fsSL "https://github.com/cowsql/cowsql/archive/refs/tags/v${COWSQL_VERSION}.tar.gz"
tar --force-local -xzvf "cowsql-${COWSQL_VERSION}.tar.gz" -C "${SYSEXTNAME}"
cd "${SYSEXTNAME}"/cowsql-${COWSQL_VERSION}
export PKG_CONFIG_PATH="${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/lib/pkgconfig"
export CPPFLAGS="-I${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/include -I${SCRIPTFOLDER}/${SYSEXTNAME}/usr/include"
export LD_LIBRARY_PATH="${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/lib"
autoreconf -i
./configure --prefix="${BLUEFINPREFIX}"
make
make install DESTDIR="${SCRIPTFOLDER}/${SYSEXTNAME}"

cd "${SCRIPTFOLDER}"

# lxc
rm -f "lxc-${LXC_VERSION}.tar.gz"
curl -o "lxc-${LXC_VERSION}.tar.gz" -fsSL "https://linuxcontainers.org/downloads/lxc/lxc-${LXC_VERSION}.tar.gz"

tar --no-same-owner --no-same-permissions --force-local -xf "lxc-${LXC_VERSION}.tar.gz" -C "${SYSEXTNAME}"
rm "lxc-${LXC_VERSION}.tar.gz"

cd "${SYSEXTNAME}"/lxc-${LXC_VERSION}

meson setup --prefix="${BLUEFINPREFIX}" build/ 
ninja -C build/

make install DESTDIR="${SCRIPTFOLDER}/${SYSEXTNAME}"

# lxcfs
cd "${SCRIPTFOLDER}"
rm -f "lxcfs-${LXCFS_VERSION}.tar.gz"
curl -o "lxcfs-${LXCFS_VERSION}.tar.gz" -fsSL "https://linuxcontainers.org/downloads/lxcfs/lxcfs-${LXCFS_VERSION}.tar.gz"

tar --no-same-owner --no-same-permissions --force-local -xf "lxcfs-${LXCFS_VERSION}.tar.gz" -C "${SYSEXTNAME}"
rm "lxcfs-${LXCFS_VERSION}.tar.gz"
cd "${SYSEXTNAME}"/lxcfs-${LXCFS_VERSION}

meson setup --prefix="${BLUEFINPREFIX}" build/ 
ninja -C build/
make install DESTDIR="${SCRIPTFOLDER}/${SYSEXTNAME}"

echo "lxcfs done"

# incus
cd "${SCRIPTFOLDER}"

rm -f "incus-${VERSION}.tar.xz"
POINTTWO=${VERSION%.*}

# source
curl -o "incus-${VERSION}.tar.xz" -fsSL "https://github.com/lxc/incus/releases/download/v${VERSION}/incus-${POINTTWO}.tar.xz"
# TODO: Also allow to consume upstream containerd and runc release binaries with their respective versions

tar --no-same-owner --no-same-permissions --force-local -xf "incus-${VERSION}.tar.xz" -C "${SYSEXTNAME}"
rm "incus-${VERSION}.tar.xz"
mkdir -p "${SYSEXTNAME}"/usr/bluefin/bin

cd "${SYSEXTNAME}"/incus-${POINTTWO}
#make deps
export CGO_CFLAGS="-I${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/include"
export CGO_LDFLAGS="-L${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/lib/"
export CGO_LDFLAGS_ALLOW="(-Wl,-wrap,pthread_create)|(-Wl,-z,now)"
make
cp "${GOPATH}/bin/incus" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/incusd" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/fuidshift" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/generate" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/incus-agent" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/incus-benchmark" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/incus-migrate" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/incus-user" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/lxc-to-incus" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"
cp "${GOPATH}/bin/lxd-to-incus" "${SCRIPTFOLDER}/${SYSEXTNAME}/usr/bluefin/bin/"

cd "${SCRIPTFOLDER}/${SYSEXTNAME}"
rm -rf cowsql-${COWSQL_VERSION}
rm -rf incus-${POINTTWO}
rm -rf pkg
rm -rf bin
rm -rf raft

#rmdir "${SYSEXTNAME}"/docker
#mkdir -p "${SYSEXTNAME}/usr/lib/systemd/system"
#if [ "${ONLY_CONTAINERD}" = 1 ]; then
#  rm "${SYSEXTNAME}/usr/bin/docker" "${SYSEXTNAME}/usr/bin/dockerd" "${SYSEXTNAME}/usr/bin/docker-init" "${SYSEXTNAME}/usr/bin/docker-proxy"
#elif [ "${ONLY_DOCKER}" = 1 ]; then
#  rm "${SYSEXTNAME}/usr/bin/containerd" "${SYSEXTNAME}/usr/bin/containerd-shim-runc-v2" "${SYSEXTNAME}/usr/bin/ctr" "${SYSEXTNAME}/usr/bin/runc"
#  if [[ "${VERSION%%.*}" -lt 23 ]] ; then
#    # Binary releases 23 and higher don't ship containerd-shim
#    rm "${SYSEXTNAME}/usr/bin/containerd-shim"
#  fi
#fi
#if [ "${ONLY_CONTAINERD}" != 1 ]; then
#  cat > "${SYSEXTNAME}/usr/lib/systemd/system/docker.socket" <<-'EOF'
#	[Unit]
#	PartOf=docker.service
#	Description=Docker Socket for the API
#	[Socket]
#	ListenStream=/var/run/docker.sock
#	SocketMode=0660
# 	SocketUser=root
# 	SocketGroup=docker
# 	[Install]
# 	WantedBy=sockets.target
# EOF
#   mkdir -p "${SYSEXTNAME}/usr/lib/systemd/system/sockets.target.d"
#   { echo "[Unit]"; echo "Upholds=docker.socket"; } > "${SYSEXTNAME}/usr/lib/systemd/system/sockets.target.d/10-docker-socket.conf"
#   cat > "${SYSEXTNAME}/usr/lib/systemd/system/docker.service" <<-'EOF'
# 	[Unit]
# 	Description=Docker Application Container Engine
# 	After=containerd.service docker.socket network-online.target
# 	Wants=network-online.target
# 	Requires=containerd.service docker.socket
# 	[Service]
# 	Type=notify
# 	EnvironmentFile=-/run/flannel/flannel_docker_opts.env
# 	Environment=DOCKER_SELINUX=--selinux-enabled=true
# 	ExecStart=/usr/bin/dockerd --host=fd:// --containerd=/run/containerd/containerd.sock $DOCKER_SELINUX $DOCKER_OPTS $DOCKER_CGROUPS $DOCKER_OPT_BIP $DOCKER_OPT_MTU $DOCKER_OPT_IPMASQ
# 	ExecReload=/bin/kill -s HUP $MAINPID
# 	LimitNOFILE=1048576
# 	# Having non-zero Limit*s causes performance problems due to accounting overhead
# 	# in the kernel. We recommend using cgroups to do container-local accounting.
# 	LimitNPROC=infinity
# 	LimitCORE=infinity
# 	# Uncomment TasksMax if your systemd version supports it.
# 	# Only systemd 226 and above support this version.
# 	TasksMax=infinity
# 	TimeoutStartSec=0
# 	# set delegate yes so that systemd does not reset the cgroups of docker containers
# 	Delegate=yes
# 	# kill only the docker process, not all processes in the cgroup
# 	KillMode=process
# 	# restart the docker process if it exits prematurely
# 	Restart=on-failure
# 	StartLimitBurst=3
# 	StartLimitInterval=60s
# 	[Install]
# 	WantedBy=multi-user.target
# EOF
# fi
# if [ "${ONLY_DOCKER}" != 1 ]; then
#   cat > "${SYSEXTNAME}/usr/lib/systemd/system/containerd.service" <<-'EOF'
# 	[Unit]
# 	Description=containerd container runtime
# 	After=network.target
# 	[Service]
# 	Delegate=yes
# 	Environment=CONTAINERD_CONFIG=/usr/share/containerd/config.toml
# 	ExecStartPre=mkdir -p /run/docker/libcontainerd
# 	ExecStartPre=ln -fs /run/containerd/containerd.sock /run/docker/libcontainerd/docker-containerd.sock
# 	ExecStart=/usr/bin/containerd --config ${CONTAINERD_CONFIG}
# 	KillMode=process
# 	Restart=always
# 	# (lack of) limits from the upstream docker service unit
# 	LimitNOFILE=1048576
# 	LimitNPROC=infinity
# 	LimitCORE=infinity
# 	TasksMax=infinity
# 	[Install]
# 	WantedBy=multi-user.target
# EOF
#   mkdir -p "${SYSEXTNAME}/usr/lib/systemd/system/multi-user.target.d"
#   { echo "[Unit]"; echo "Upholds=containerd.service"; } > "${SYSEXTNAME}/usr/lib/systemd/system/multi-user.target.d/10-containerd-service.conf"
#   mkdir -p "${SYSEXTNAME}/usr/share/containerd"
#   cat > "${SYSEXTNAME}/usr/share/containerd/config.toml" <<-'EOF'
# 	version = 2
# 	# set containerd's OOM score
# 	oom_score = -999
# 	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
# 	# setting runc.options unsets parent settings
# 	runtime_type = "io.containerd.runc.v2"
# 	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
# 	SystemdCgroup = true
# EOF
#   sed 's/SystemdCgroup = true/SystemdCgroup = false/g' "${SYSEXTNAME}/usr/share/containerd/config.toml" > "${SYSEXTNAME}/usr/share/containerd/config-cgroupfs.toml"
# fi
