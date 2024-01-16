wasmtimeversion := "13.0.0"
neovimversion := "0.9.5"
goversion := "1.21.6"

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


# install go
go: build-go systemd-sysext
    #!/usr/bin/env bash
    echo "Installing go extension, requires elevated permissions"
    sudo cp result/go.raw /var/lib/extensions/go.raw
    echo "Reloading system extensions, requires elevated permissions"
    echo "Add /usr/local/go/bin to your PATH to use go"
    sudo systemd-sysext refresh
    systemd-sysext

# install caddy
caddy: build-caddy systemd-sysext
    #!/usr/bin/env bash
    echo "Installing caddy extension, requires elevated permissions"
    sudo cp result/caddy.raw /var/lib/extensions/caddy.raw
    echo "Reloading system extensions, requires elevated permissions"
    sudo systemd-sysext refresh
    systemd-sysext

# install albafetch from flakehub reference
albafetch: (container "nix") (build-nix "https://flakehub.com/f/alba4k/albafetch/0.1.570.tar.gz" "albafetch") systemd-sysext
    #!/usr/bin/env bash
    echo "Installing albafetch extension, requires elevated permissions"
    sudo cp result/albafetch.raw /var/lib/extensions/albafetch.raw
    echo "Reloading system extensions, requires elevated permissions"
    sudo systemd-sysext refresh
    systemd-sysext

[private]
build-neovim: (container "neovim")
    @echo "Building neovim sysext"
    @podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/neovim:latest /bakery/create_neovim_sysext.sh {{neovimversion}} neovim >/dev/null

[private]
build-caddy: (container "caddy")
    @echo "Building caddy sysext"
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/caddy:latest /bakery/create_caddy_sysext.sh latest caddy


[private]
build-go: (container "go")
    @echo "Building go sysext"
    @podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/go:latest /bakery/create_go_sysext.sh {{goversion}} go >/dev/null

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
build-nix FLAKEREF PACKAGE: (container "nix")
    podman run --rm -e OS=_any -v `pwd`/result:/bakery/result ${USER}/nix:latest /bakery/nix_bundle_sysext.sh {{FLAKEREF}} {{PACKAGE}} 


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
