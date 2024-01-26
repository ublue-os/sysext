_default:
  @just --list

remove NAME:
    @echo "Removing {{NAME}} extension, requires elevated permissions"
    sudo rm -f /var/lib/extensions/{{NAME}}.raw
    sudo systemd-sysext refresh 
    @just refresh-store
    systemd-sysext

build-config OUT_DIR:
    #!/usr/bin/env bash
    set -euo pipefail
    CONTAINER_MANAGER="${CONTAINER_MANAGER-podman}"

    if [ ! type "$CONTAINER_MANAGER" &>/dev/null ] ; then
      echo "Failed running $CONTAINER_MANAGER"
      exit 1
    fi

    "$CONTAINER_MANAGER" run --rm -v ${PWD}:/app:Z -w /app docker.io/nixos/nix:latest nix run --extra-experimental-features nix-command --extra-experimental-features flakes .#compile-configuration "{{OUT_DIR}}"

add-overlay FILE_PATH: systemd-sysext setup-nix-mount 
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing provided system extension"
    sudo mkdir -p /var/lib/extensions
    sudo cp "{{FILE_PATH}}" /var/lib/extensions
    echo "Reloading system extensions"
    @just merge-overlays 
    systemd-sysext

merge-overlays:
    sudo systemd-sysext merge
    @just refresh-store

[private]
refresh-store:
    if [ -e /nix ] ; then
    sudo umount /nix/store
    sudo mount --bind /usr/store /nix/store
    fi

[private]
setup-nix-mount:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -e /nix ] ; then
      echo "Creating /nix/store"
      sudo chattr -i /
      sudo mkdir -p /nix/store
      sudo chattr +i /
    fi
    if [ ! -e /tmp/nix-store-bindmount ] ; then
      echo "Creating /nix/store bind-mount"
      sudo mkdir -p /tmp/nix-store-bindmount
      sudo mount --bind /nix/store /tmp/nix-store-bindmount
      sudo mount --bind /tmp/nix-store-bindmount /nix/store
    fi

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
