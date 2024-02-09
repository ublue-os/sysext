{
  description = "Systemd System Extensions using the Nix Store";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
        lib = pkgs.lib;

        impureOnlyDerivation = pkgs.stdenv.mkDerivation {
          name = "invalid-impure-derivation";
          buildCommand = ''
            ${builtins.abort "This package is only available in Impure mode only"}
          '';
        };
        bext_deps = {
          build = with pkgs; [go pkg-config];
          runtime = with pkgs; [btrfs-progs gpgme lvm2];
        };
      in {
        formatter = pkgs.alejandra;
        devShells.default = pkgs.mkShell.override {stdenv = pkgs.llvmPackages_14.stdenv;} {
          packages = with pkgs; [cobra-cli gopls eclint] ++ bext_deps.build ++ bext_deps.runtime;
        };
        packages = {
          default = self.packages.${system}.bext;

          bext = pkgs.buildGoModule.override {stdenv = pkgs.llvmPackages_14.stdenv;} {
            pname = "sysext";
            name = "bext";
            src = ./.;
            pwd = ./.;
            nativeBuildInputs = bext_deps.build;
            buildInputs = bext_deps.runtime;
            vendorHash = "sha256-tJMrrMWLcfYstD9I1poDpT5MX75066k7hUjfTDCv/i4=";
          };

          bextStatic = self.packages.${system}.bext.overrideAttrs (final: oldAttrs: {
            nativeBuildInputs = [pkgs.musl] ++ oldAttrs.nativeBuildInputs;

            LDFLAGS = [
              "-static"
              "-L${pkgs.musl}/lib"
              "-L${pkgs.gpgme}/lib"
              "-L${pkgs.lvm2}/lib"
            ];

            ldflags = [
              "-linkmode external"
            ];
          });

          bake-recipe =
            if lib.inPureEvalMode
            then impureOnlyDerivation
            else let
              config-envvar = "BEXT_CONFIG_FILE";
              config =
                pkgs.lib.trivial.importJSON (/. + builtins.getEnv config-envvar);
              all_deps =
                builtins.map (package: pkgs.${package}) config.packages;

              generate-recipe-derivation = pkgs.symlinkJoin {
                name = "derivation-from-recipe";
                paths = all_deps;
              };

              bundle-recipe-derivations = pkgs.stdenv.mkDerivation {
                name = config.sysext-name + "-store";
                buildInputs = with pkgs; [perl gnutar];
                exportReferencesGraph = lib.lists.flatten (builtins.map (x: [("closure-" + baseNameOf x) x]) all_deps);
                buildCommand = ''
                  storePaths=$(${lib.getExe pkgs.perl} ${pkgs.pathsFromGraph} ./closure-*)

                  mkdir $out
                  ${lib.getExe pkgs.rsync} -a $storePaths $out
                '';
              };
            in
              pkgs.stdenv.mkDerivation {
                name = config.sysext-name + "-baked";
                buildInputs = with pkgs; [coreutils squashfsTools];
                buildCommand = ''
                  set -eoux pipefail

                  mkdir -p usr/{store,lib/extension-release.d,extensions.d/${config.sysext-name}/bin,}
                  cp -R -u -v ${generate-recipe-derivation}/* usr &

                  {
                    echo "ID=${config.os}"
                    echo "EXTENSION_RELOAD_MANAGER=1"
                    if [ "${config.os}" != "_any" ]; then
                      echo "SYSEXT_LEVEL=1.0"
                    fi
                    if [ "${config.arch}" != "" ]; then
                      echo "ARCHITECTURE=${config.arch}"
                    fi
                  } > "usr/lib/extension-release.d/extension-release.${config.sysext-name}.sysext" &

                  echo '${builtins.toJSON config}' > usr/extensions.d/${config.sysext-name}/metadata.json &

                  # Upstream Issue: https://github.com/NixOS/nixpkgs/issues/252620
                  #{
                  #  $pkgs.selinux-python/bin/semanage fcontext -a -t etc_t 'usr/etc(/.*)?'
                  #  $pkgs.selinux-python/bin/semanage fcontext -a -t lib_t 'usr/lib(/.*)?'
                  #  $pkgs.selinux-python/bin/semanage fcontext -a -t man_t 'usr/man(/.*)?'
                  #  $pkgs.selinux-python/bin/semanage fcontext -a -t bin_t 'usr/s?bin(/.*)?'
                  #  $pkgs.selinux-python/bin/semanage fcontext -a -t usr_t 'usr/share(/.*)?'
                  #  $pkgs.selinux-python/bin/restorecon -Rv *
                  #} &

                  cp -r ${bundle-recipe-derivations}/* usr/store &

                  wait $(jobs -p)

                  chmod -R 755 usr

                  {
                    mv usr/bin/* usr/extensions.d/${config.sysext-name}/bin
                    rm -r usr/bin
                  }

                  shopt -s extglob
                  rm -- !(usr)
                  ${pkgs.squashfsTools}/bin/mksquashfs \
                    . \
                    $out \
                    -root-mode 755 -all-root -no-hardlinks -exit-on-error -progress -action "chmod(755)@true"
                '';
              };
        };
      }
    );
}
