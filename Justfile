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
    podman run --rm -w /app -v ${PWD}:/app:Z nixos/nix:latest nix run --extra-experimental-features nix-command --extra-experimental-features flakes .#compile-configuration "{{OUT_DIR}}"

set-overlay FILE_PATH: systemd-sysext 
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing provided system extension, requires elevated permissions"
    sudo mkdir -p /var/lib/extensions
    sudo cp "{{FILE_PATH}}" /var/lib/extensions
    echo "Reloading system extensions, requires elevated permissions"
    sudo systemd-sysext refresh 
    systemd-sysext

mount-store-squashfs FILE_PATH: (setup-nix-mount)
    #!/usr/bin/env bash
    set -euo pipefail
    mount -t squashfs -o loop "{{FILE_PATH}}" /nix/store

[private]
setup-nix-mount:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -e /nix ] ; then
      chattr +i /
      mkdir /nix
      chattr -i /
    fi
    if [ ! -e /tmp/nix-store-bindmount ] ; then
      mkdir -p /tmp/nix-store-bindmount
      mount --bind /nix/store /tmp/nix-store-bindmount
      mount --bind /tmp/nix-store-bindmount /nix/store
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
