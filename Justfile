incus-container:
    podman build -t ${USER}/incus:latest -f incus/Containerfile.incus .

build-incus: incus-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/incus:latest /bakery/create_incus_sysext.sh 0.4.0 incus

incus: build-incus
    echo "installing incus"

docker-container:
    podman build -t ${USER}/docker:latest -f docker/Containerfile.docker .

docker: build-docker ensure-systemd-sysext
    echo "installing docker"

build-docker: docker-container
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/docker:latest /bakery/create_docker_sysext.sh 24.0.6 docker

dockercompose: 
    ./create_docker_compose_sysext.sh 2.22.0 dockercompose

systemd-sysext:
    sudo systemctl enable --now systemd-sysext.service

ensure-systemd-sysext: systemd-sysext
    sudo cp systemd/ensure-sysext.service /etc/systemd/system/ensure-sysext.service
    sudo systemctl daemon-reload
    sudo systemctl enable ensure-sysext.service