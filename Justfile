incus-container:
    podman build -t ${USER}/incus:latest -f builders/incus/Containerfile.incus .

build-incus: incus-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/incus:latest /bakery/create_incus_sysext.sh 0.4.0 ublueincus

incus: build-incus
    echo "installing incus"

docker-container:
    podman build -t ${USER}/docker:latest -f builders/docker/Containerfile.docker .

docker: build-docker ensure-systemd-sysext
    echo "installing docker"
    sudo mkdir -p /etc/extensions
    sudo cp result/ubluedocker.raw /etc/extensions/ubluedocker.raw
    echo "Reboot to enable docker"

build-docker: docker-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/docker:latest /bakery/create_docker_sysext.sh 24.0.6 ubluedocker

build-docker-compose: docker-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/docker:latest /bakery/create_docker_compose_sysext.sh 2.22.0 ubluedockercompose


build-docker-local: 
    export OS=_any; ./create_docker_sysext.sh 24.0.6 ubluedocker

build-udocker-local: 
    export OS=_any; ./create_udocker_portable.sh 24.0.6 udocker

build-docker-compose-local: 
    export OS=_any; ./create_docker_compose_sysext.sh 2.22.0 ubluedockercompose

dockercompose: 
    ./create_docker_compose_sysext.sh 2.22.0 dockercompose

wasmtime-container:
    podman build -t ${USER}/wasmtime:latest -f builders/wasmtime/Containerfile.wasmtime .

build-wasmtime: wasmtime-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/wasmtime:latest /bakery/create_wasmtime_sysext.sh 13.0.0 wasmtime



build-wasmtime-local: 
    export OS=_any; ./create_wasmtime_sysext.sh 13.0.0 wasmtime


systemd-sysext:
    sudo systemctl enable --now systemd-sysext.service

ensure-systemd-sysext: systemd-sysext
    sudo cp systemd/ensure-sysext.service /etc/systemd/system/ensure-sysext.service
    sudo systemctl daemon-reload
    sudo systemctl enable ensure-sysext.service