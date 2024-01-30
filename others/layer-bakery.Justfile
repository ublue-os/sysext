_default:
  @just --list

enable-sysext-support:
  sudo setenforce 0 || true

disable-sysext-support:
  echo "Disabling sysext support requires the layers to me unmerged and SELinux will be turned on again. Please do not merge any layers while SELinux is enabled orelse your system will break!"
  sudo systemd-sysext unmerge
  sudo setenforce 1

build-config CONFIG_FILE:
    #!/usr/bin/env bash
    set -euo pipefail
    CONTAINER_MANAGER="${CONTAINER_MANAGER-podman}"

    if [ ! type "$CONTAINER_MANAGER" &>/dev/null ] ; then
      echo "Failed running $CONTAINER_MANAGER"
      exit 1
    fi

    "$CONTAINER_MANAGER" run --rm -v ${PWD}/..:/app:Z -w /app --mount type=bind,source=$(realpath {{CONFIG_FILE}}),target=/config.json,readonly docker.io/nixos/nix:latest sh -c "BEXT_CONFIG_FILE=/config.json nix build --extra-experimental-features nix-command --extra-experimental-features flakes --impure .#bake-recipe && cp result ./layer_result.sysext.raw"
    mv ./layer_result.sysext.raw $(jq '."sysext-name"' $(realpath {{CONFIG_FILE}})).sysext.raw

add-overlay FILE_PATH: systemd-sysext enable-sysext-support setup-nix-mount 
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing provided system extension"
    sudo mkdir -p /var/lib/extensions
    sudo cp "{{FILE_PATH}}" /var/lib/extensions
    echo "Reloading system extensions"
    just merge-overlays
    just update-path
    systemd-sysext
    echo "Make sure that /tmp/extensions.d/bin is in your PATH variable."

merge-overlays: enable-sysext-support 
    sudo systemd-sysext refresh 
    @just refresh-store
    @just update-path

remove NAME:
    @echo "Removing {{NAME}} extension, requires elevated permissions"
    sudo rm -f /var/lib/extensions/{{NAME}}.raw
    sudo systemd-sysext refresh
    @just refresh-store
    sudo umount /tmp/extensions.d/bin &> /dev/null || true
    @just update-path
    systemd-sysext

[private]
update-path:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -e /usr/extensions.d/ ] ; then 
      exit 1
    fi
    sudo umount /tmp/extensions.d/bin &> /dev/null || true
    sudo mkdir -p /tmp/extensions.d/bin

    # 3 here because find prints the directory name itself too
    if [ $(find /usr/extensions.d/ -maxdepth 1 | wc -l) -lt 3 ] ; then
      sudo mount --bind /usr/extensions.d/*/bin /tmp/extensions.d/bin
    else
      sudo mount -t overlay -o lowerdir=$(for PATH_ENV in /usr/extensions.d/*; do echo -n "$PATH_ENV/bin:"; done | sed 's/:$//') none /tmp/extensions.d/bin
    fi

[private]
refresh-store:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -e /nix ] ; then
      sudo umount /nix/store
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
    sudo umount /tmp/nix-store-bindmount &> /dev/null || true
    sudo umount /nix/store &> /dev/null || true
    sudo umount /tmp/nix-store-bindmount &> /dev/null || true
    sudo rm -rf /tmp/nix-store-bindmount
    sudo umount /tmp/extensions.d/bin &> /dev/null || true
    sudo rm -rf /tmp/extensions.d/

[private]
systemd-sysext:
    #!/usr/bin/env bash
    systemctl --quiet is-enabled systemd-sysext || { echo "enabling systemd-sysext"; sudo systemctl enable --now systemd-sysext.service; }
    test -d /var/lib/extensions || { echo "creating /var/lib/extensions"; sudo mkdir -p /var/lib/extensions; }
    systemctl --quiet is-enabled systemd-confext || { echo "enabling systemd-confext"; sudo systemctl enable --now systemd-confext.service; }
    test -d /var/lib/confexts || { echo "creating /var/lib/confexts"; sudo mkdir -p /var/lib/confexts; }
