{
  description = "Systemd System Extensions using the Nix Store";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-23.11";
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
        config-envvar = "BEXT_CONFIG_FILE";
        config = pkgs.lib.trivial.importJSON (/. + builtins.getEnv config-envvar);
        all_deps = builtins.map (package: pkgs.${package}) config.packages;
      in {
        formatter = pkgs.alejandra;
        devShells.default = pkgs.mkShell.override { stdenv = pkgs.llvmPackages_14.stdenv; } ({
          packages = with pkgs; [cobra-cli go gopls eclint pkg-config btrfs-progs gpgme lvm2];
        });
        packages = {
          default = self.packages.${system}.bake-recipe;

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

          generate-recipe-derivation = pkgs.symlinkJoin {
            name = "derivation-from-recipe";
            paths = all_deps;
            postBuild = ''
              rm -rf "$out/nix-support"
            '';
          };

          bake-recipe = pkgs.stdenv.mkDerivation {
            name = config.sysext-name + "-baked";
            buildInputs = with pkgs; [coreutils squashfsTools];
            buildCommand = ''
              set -eoux pipefail

              mkdir usr
              cp -r ${self.packages.${system}.generate-recipe-derivation}/* usr
              chmod -R 755 usr
              
              {
                mkdir -p usr/lib/extension-release.d
                {
                  echo "ID=${config.os}"
                  echo "EXTENSION_RELOAD_MANAGER=1"
                  if [ "${config.os}" != "_any" ]; then
                    echo "SYSEXT_LEVEL=1.0"
                  fi
                  if [ "${config.arch}" != "" ]; then
                    echo "ARCHITECTURE=${config.arch}"
                  fi
                } > "usr/lib/extension-release.d/extension-release.${config.sysext-name}.sysext"
              } &

              mkdir -p "usr/extensions.d/${config.sysext-name}/bin"
              {
                mv usr/bin/* "usr/extensions.d/${config.sysext-name}/bin"
                rm -r usr/bin
              } &

              echo '${builtins.toJSON config}' | tee usr/extensions.d/${config.sysext-name}/metadata.json
              
              mkdir -p usr/store
              cp -r ${(self.packages.${system}.bundle-recipe-derivations)}/* ./usr/store

              # Upstream Issue: https://github.com/NixOS/nixpkgs/issues/252620
              #$pkgs.selinux-python/bin/semanage fcontext -a -t etc_t 'usr/etc(/.*)?' &
              #$pkgs.selinux-python/bin/semanage fcontext -a -t lib_t 'usr/lib(/.*)?' &
              #$pkgs.selinux-python/bin/semanage fcontext -a -t man_t 'usr/man(/.*)?' &
              #$pkgs.selinux-python/bin/semanage fcontext -a -t bin_t 'usr/s?bin(/.*)?' &
              #$pkgs.selinux-python/bin/semanage fcontext -a -t usr_t 'usr/share(/.*)?' &
              #$pkgs.selinux-python/bin/restorecon -Rv *

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
