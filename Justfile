_default:
  @just --list


remove NAME:
    @echo "Removing {{NAME}} extension, requires elevated permissions"
    sudo rm -f /var/lib/extensions/{{NAME}}.raw
    sudo systemd-sysext refresh 
    systemd-sysext

build-config OUT_DIR:
    #!/usr/bin/env bash
    set -euo pipefail
    podman run --rm -w /app -v ${PWD}:/app nixos/nix:latest nix run --extra-experimental-features nix-command --extra-experimental-features flakes .#compile-configuration "{{OUT_DIR}}"

[private]
container NAME:
    @echo "Building {{NAME}} container"
    @podman build -t ${USER}/{{NAME}}:latest -f builders/{{NAME}}/Containerfile.{{NAME}} . >/dev/null

[private]
clean:
    @rm -rf result/*.raw

[private]
systemd-sysext:
    #!/usr/bin/env bash
    systemctl --quiet is-enabled systemd-sysext || { echo "enabling systemd-sysext"; sudo systemctl enable --now systemd-sysext.service; }
    test -d /var/lib/extensions || { echo "creating /var/lib/extensions"; sudo mkdir -p /var/lib/extensions; }
    systemctl --quiet is-enabled systemd-confext || { echo "enabling systemd-confext"; sudo systemctl enable --now systemd-confext.service; }
    test -d /var/lib/confexts || { echo "creating /var/lib/confexts"; sudo mkdir -p /var/lib/confexts; }
