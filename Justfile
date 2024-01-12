wasmtimeversion := "13.0.0"
neovimversion := "0.9.5"

_default:
  @just --list

[private]
build-wasmtime: (container "wasmtime")
    @echo "Building wasmtime sysext"
    @podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/wasmtime:latest /bakery/create_wasmtime_sysext.sh {{wasmtimeversion}} wasmtime >/dev/null

# install wasmtime
wasmtime: build-wasmtime systemd-sysext
    #!/usr/bin/env bash
    echo "Installing wasmtime extension, requires elevated permissions"
    sudo cp result/wasmtime.raw /var/lib/extensions/wasmtime.raw
    echo "Reloading system extensions, requires elevated permissions"
    sudo systemd-sysext refresh
    systemd-sysext

# install neovim
neovim: build-neovim systemd-sysext
    #!/usr/bin/env bash
    echo "Installing neovim extension, requires elevated permissions"
    sudo cp result/neovim.raw /var/lib/extensions/neovim.raw
    echo "Reloading system extensions, requires elevated permissions"
    sudo systemd-sysext refresh
    systemd-sysext

# install vscode
vscode: build-vscode systemd-sysext
    #!/usr/bin/env bash
    echo "Installing vscode extension, requires elevated permissions"
    sudo cp result/vscode.raw /var/lib/extensions/vscode.raw
    echo "Reloading system extensions, requires elevated permissions"
    sudo systemd-sysext refresh
    systemd-sysext

[private]
build-neovim: (container "neovim")
    @echo "Building neovim sysext"
    @podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/neovim:latest /bakery/create_neovim_sysext.sh {{neovimversion}} neovim >/dev/null

[private]
build-vscode: (container "vscode")
    @echo "Building vscode sysext"
    @podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/vscode:latest /bakery/create_vscode_sysext.sh latest vscode >/dev/null

# remove an installed extension by name `just remove vscode`
remove NAME:
    @echo "Removing {{NAME}} extension, requires elevated permissions"
    sudo rm -f /var/lib/extensions/{{NAME}}.raw
    sudo systemd-sysext refresh 
    systemd-sysext

[private]
container NAME:
    @echo "Building {{NAME}} container"
    @podman build -t ${USER}/{{NAME}}:latest -f builders/{{NAME}}/Containerfile.{{NAME}} . >/dev/null

[private]
local NAME VERSION:
    @echo "Building {{NAME}} locally"
    export OS=_any; ./create_{{NAME}}_sysext.sh {{VERSION}} {{NAME}}

[private]
clean:
    @rm -rf result/*.raw

[private]
build-wasmtime-local: 
    export OS=_any; ./create_wasmtime_sysext.sh 13.0.0 wasmtime

[private]
systemd-sysext:
    #!/usr/bin/env bash
    systemctl --quiet is-enabled systemd-sysext || { echo "enabling systemd-sysext"; sudo systemctl enable --now systemd-sysext.service; }
    test -d /var/lib/extensions || { echo "creating /var/lib/extensions"; sudo mkdir -p /var/lib/extensions; }
