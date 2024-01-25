{
  description = "Systemd System Extensions using the Nix Store";

  inputs = {
    nixpkgs.url = "nixpkgs/release-23.11";
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
        config = pkgs.lib.trivial.importJSON ./config.json;
        all_deps = builtins.map (package: pkgs.${package}) config.packages;
        #++ (lib.lists.flatten (builtins.map (package: pkgs.${package}.buildInputs) config.packages));
        squashfs-build = ''
          set -euo pipefail
          storePaths=$(${lib.getExe pkgs.perl} ${pkgs.pathsFromGraph} ./closure-*)

          # 64 cores on i686 does not work
          # fails with FATAL ERROR: mangle2:: xz compress failed with error code 5
          if [[ "$NIX_BUILD_CORES" -gt 48 ]] ; then 
            NIX_BUILD_CORES=48
          fi
          
          ${pkgs.squashfsTools}/bin/mksquashfs \
            $storePaths \
            $out \
            -all-root -no-hardlinks -exit-on-error -info -progress -processors $NIX_BUILD_CORES
        '';
      in {
        formatter = pkgs.alejandra;
        bundlers = {
          toTar = {...} @ drv:
            pkgs.stdenv.mkDerivation {
              name = drv.pname + "-store.sqfs";
              buildInputs = [pkgs.perl pkgs.gnutar];
              exportReferencesGraph = lib.lists.flatten (builtins.map (x: [("closure-" + baseNameOf x) x]) (lib.lists.flatten (pkgs.${drv.pname})));
              buildCommand = squashfs-build;
            };
        };
        packages = {
          default = self.packages.${system}.compile-configuration;
          bundle-all-config = pkgs.stdenv.mkDerivation {
            name = config.sysext-name + "-store.sqfs";
            buildInputs = [pkgs.perl pkgs.gnutar];
            exportReferencesGraph = lib.lists.flatten (builtins.map (x: [("closure-" + baseNameOf x) x]) all_deps);
            buildCommand = squashfs-build;
          };

          sysext-derivation-from-config = pkgs.symlinkJoin {
            name = "derivation-with-config.json-outputs";
            paths = all_deps;
            meta.priority = 10;
            postBuild = ''
              rm $out/nix-support -rf
              mkdir -p $out/usr
              ${lib.getExe pkgs.findutils} $out -maxdepth 1 -type d -exec mv {} $out/usr/ \;
            '';
          };

          sysext-image-maker = pkgs.writeShellScriptBin "makeSysext.sh" ''
            # got this script from flatcar linux bakery!
            FORMAT="$1"
            DIRECTORY="$2"
            TARGET_NAME="$3"
            if [ $# -lt 3 ]; then
              echo "Usage: $0 FORMAT DIRECTORY"
              echo "The script will make a TARGET_NAME.raw image of the folder DIRECTORY with FORMAT format."
              exit 1
            fi
            if [ "$FORMAT" != "squashfs" ] && [ "$FORMAT" != "btrfs" ] && [ "$FORMAT" != "ext4" ] && [ "$FORMAT" != "ext2" ]; then
              echo "Expected FORMAT=squashfs, FORMAT=btrfs, FORMAT=ext4, or FORMAT=ext2, got '$FORMAT'" >&2
            exit 1
            fi

            if [ "$FORMAT" = "btrfs" ]; then
              # Note: We didn't chown to root:root, meaning that the file ownership is left as is
              ${pkgs.btrfs-progs}/bin/mkfs.btrfs --mixed -m single -d single --shrink --rootdir "$DIRECTORY" "$DIRECTORY".raw
              # This is for testing purposes and makes not much sense to use because --rootdir doesn't allow to enable compression
            elif [ "$FORMAT" = "ext4" ] || [ "$FORMAT" = "ext2" ]; then
              # Assuming that 1 GB is enough
              ${pkgs.coreutils}/bin/truncate -s 1G "$DIRECTORY".raw
              # Note: We didn't chown to root:root, meaning that the file ownership is left as is
              ${pkgs.e2fsprogs}/bin/mkfs."$FORMAT" -E root_owner=0:0 -d "$DIRECTORY" "$DIRECTORY".raw
              ${pkgs.e2fsprogs}/bin/resize2fs -M "$DIRECTORY".raw
            else
              ${pkgs.squashfsTools}/bin/mksquashfs "$DIRECTORY" "$TARGET_NAME".raw -all-root
            fi
          '';

          compile-configuration = pkgs.writeShellScriptBin "compiler.sh" ''
            set -euox pipefail

            OUT_DIR="$1"
            if [ "$OUT_DIR" != \"\" ] ; then
              mkdir -p "$OUT_DIR"
              OUT_DIR="$OUT_DIR/"
              pushd $OUT_DIR
              trap popd EXIT
            fi

            ${lib.getExe self.packages.${system}.sysext-image-maker} \
              squashfs \
              ${self.packages.${system}.sysext-derivation-from-config} \
              "${config.sysext-name}.sysext"

            # Modify sysext to have valid structure

            mkdir -p squashfs-root
            ${pkgs.squashfsTools}/bin/unsquashfs -d squashfs-root ${config.sysext-name}.sysext.raw 
            rm -f ${config.sysext-name}.sysext.raw 
            
            chmod -R 755 squashfs-root 
            chmod -R 755 squashfs-root/* 

            pushd squashfs-root
              # Generate metadata for valid system extension
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

              mkdir -p usr/extensions.d/${config.sysext-name}/bin
              mv usr/bin/* usr/extensions.d/${config.sysext-name}/bin
              rm -r usr/bin

              # This layer's nix-store inside /usr/store
              mkdir -p usr/store
              ${pkgs.squashfsTools}/bin/unsquashfs -d ./usr/store ${(self.packages.${system}.bundle-all-config)}

              # Upstream Issue: https://github.com/NixOS/nixpkgs/issues/252620
              #$pkgs.policycoreutils/bin/semanage fcontext -a -t etc_t 'usr/etc(/.*)?'
              #$pkgs.policycoreutils/bin/semanage fcontext -a -t lib_t 'usr/lib(/.*)?'
              #$pkgs.policycoreutils/bin/semanage fcontext -a -t man_t 'usr/man(/.*)?'
              #$pkgs.policycoreutils/bin/semanage fcontext -a -t bin_t 'usr/s?bin(/.*)?'
              #$pkgs.policycoreutils/bin/semanage fcontext -a -t usr_t 'usr/share(/.*)?'
              #$pkgs.policycoreutils/bin/restorecon -Rv *

              ${pkgs.squashfsTools}/bin/mksquashfs \
                . \
                ../${config.sysext-name}.sysext.raw \
                -root-mode 755 -all-root -no-hardlinks -exit-on-error -progress -action "chmod(755)@true"
                
            popd
            
            echo "Removing remaining files"
            sudo rm -rf squashfs-root
          '';
        };
      }
    );
}
