
[private]
incus-container:
    podman build -t ${USER}/incus:latest -f builders/incus/Containerfile.incus .

[private]
build-incus: incus-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/incus:latest /bakery/create_incus_sysext.sh 0.4.0 ublueincus

[private]
incus: build-incus
    echo "installing incus"

[private]
docker-container:
    podman build -t ${USER}/docker:latest -f builders/docker/Containerfile.docker .

[private]
docker: build-docker ensure-systemd-sysext
    echo "installing docker"
    sudo mkdir -p /etc/extensions
    sudo cp result/ubluedocker.raw /etc/extensions/ubluedocker.raw
    echo "Reboot to enable docker"

[private]
build-docker: docker-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/docker:latest /bakery/create_docker_sysext.sh 24.0.6 ubluedocker

[private]
build-docker-compose: docker-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/docker:latest /bakery/create_docker_compose_sysext.sh 2.22.0 ubluedockercompose


[private]
build-docker-local: 
    export OS=_any; ./create_docker_sysext.sh 24.0.6 ubluedocker

[private]
build-udocker-local: 
    export OS=_any; ./create_udocker_portable.sh 24.0.6 udocker

[private]
build-docker-compose-local: 
    export OS=_any; ./create_docker_compose_sysext.sh 2.22.0 ubluedockercompose

[private]
dockercompose: 
    ./create_docker_compose_sysext.sh 2.22.0 dockercompose

[private]
wasmtime-container:
    podman build -t ${USER}/wasmtime:latest -f builders/wasmtime/Containerfile.wasmtime .

[private]
build-wasmtime: wasmtime-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/wasmtime:latest /bakery/create_wasmtime_sysext.sh 13.0.0 wasmtime

wasmtime: build-wasmtime systemd-sysext
    echo "Installing wasmtime"


build-wasmtime-local: 
    export OS=_any; ./create_wasmtime_sysext.sh 13.0.0 wasmtime


systemd-sysext:
    systemctl --quiet is-enabled systemd-sysext || { echo "enabling systemd-sysext"; sudo systemctl enable --now systemd-sysext.service; }
    test -d /var/lib/extensions || { echo "creating /var/lib/extensions"; sudo mkdir -p /var/lib/extensions; }
