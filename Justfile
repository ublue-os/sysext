_default:
  @just --list

enable-sysext-support:
  sudo setenforce 0

disable-sysext-support:
  echo "Disabling sysext support requires the layers to me unmerged and SELinux will be turned on again. Please do not merge any layers while SELinux is enabled orelse your system will break!"
  sudo systemd-sysext unmerge
  sudo setenforce 1

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
    just merge-overlays
    just update-path
    echo "Make sure that /run/extensions/bin is in your PATH variable."
    systemd-sysext

merge-overlays:
    sudo systemd-sysext merge
    @just refresh-store

remove NAME:
    @echo "Removing {{NAME}} extension, requires elevated permissions"
    sudo rm -f /var/lib/extensions/{{NAME}}.raw
    sudo systemd-sysext refresh
    @just refresh-store
    sudo umount /run/extensions/bin || true
    @just update-path
    systemd-sysext

[private]
unmount-store:
    sudo umount /tmp/nix-store-bindmount || true
    sudo umount /nix/store || true
    sudo rm -rf /tmp/nix-store-bindmount

[private]
update-path:
    #!/usr/bin/env bash
    set -euo pipefail
    [ ! -e /usr/extensions.d/* ] && exit 1
    sudo mkdir -p /run/extensions/bin
    for PATH_ENV in /usr/extensions.d/* ; do
      sudo mount --bind $PATH_ENV/bin /run/extensions/bin
    done

[private]
refresh-store:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -e /nix ] ; then
      sudo umount /nix/store || true
      just setup-nix-mount
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

clean:
    sudo systemd-sysext unmerge
    sudo rm -f /var/lib/extensions/*
    just unmount-store

[private]
systemd-sysext:
    #!/usr/bin/env bash
    systemctl --quiet is-enabled systemd-sysext || { echo "enabling systemd-sysext"; sudo systemctl enable --now systemd-sysext.service; }
    test -d /var/lib/extensions || { echo "creating /var/lib/extensions"; sudo mkdir -p /var/lib/extensions; }
    systemctl --quiet is-enabled systemd-confext || { echo "enabling systemd-confext"; sudo systemctl enable --now systemd-confext.service; }
    test -d /var/lib/confexts || { echo "creating /var/lib/confexts"; sudo mkdir -p /var/lib/confexts; }
