{
  description = "A very basic flake";

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
        config = pkgs.lib.trivial.importJSON ./config.json;
        wrapScript = name: pkgs.writeShellScriptBin name (builtins.readFile ./${name});
        all_deps =
          (builtins.map (package: pkgs.${package}) config.packages)
          ++ (lib.lists.flatten (builtins.map (package: pkgs.${package}.buildInputs) config.packages));
        tarfile-buildscript = ''
          storePaths=$(${lib.getExe pkgs.perl} ${pkgs.pathsFromGraph} ./closure-*)

          tar -cf - \
            --owner=0 --group=0 --mode=u+rw,uga+r \
            --hard-dereference \
            $storePaths > $out
        '';
      in {
        formatter = pkgs.alejandra;
        bundlers = {
          toTar = {...} @ drv:
            pkgs.stdenv.mkDerivation {
              name = drv.pname + "-store.tar";
              buildInputs = [pkgs.perl pkgs.gnutar];
              exportReferencesGraph = lib.lists.flatten (builtins.map (x: [("closure-" + baseNameOf x) x]) (lib.lists.flatten (pkgs.${drv.pname})));
              buildCommand = tarfile-buildscript;
            };

          bundle-all-config = _:
            pkgs.stdenv.mkDerivation {
              name = config.sysext-name + "-store.tar";
              buildInputs = [pkgs.perl pkgs.gnutar];
              exportReferencesGraph = lib.lists.flatten (builtins.map (x: [("closure-" + baseNameOf x) x]) all_deps);
              buildCommand = tarfile-buildscript;
            };
        };
        packages = {
          default = self.packages.${system}.symlinks-for-sysext;
          symlinks-for-sysext = pkgs.symlinkJoin {
            name = "derivation-with-config.json-outputs";
            paths = all_deps;
            meta.priority = 10;
            postBuild = ''
              rm $out/nix-support -rf
              mkdir -p $out/lib/extension-release.d
              {
                echo "ID=${config.os}"
                echo "EXTENSION_RELOAD_MANAGER=1"
                if [ "${config.os}" != "_any" ]; then
                  echo "SYSEXT_LEVEL=1.0"
                fi
                if [ "${config.arch}" != "" ]; then
                  echo "ARCHITECTURE=${config.arch}"
                fi
              } > "$out/lib/extension-release.d/extension-release.${config.sysext-name}.sysext"
            '';
          };
          squashfs-maker = pkgs.symlinkJoin {
            name = "sysext.sh";
            paths = [(wrapScript "sysext.sh")] ++ (with pkgs; [squashfsTools btrfs-progs]);
            buildInputs = [pkgs.makeWrapper];
            postBuild = "wrapProgram $out/bin/sysext.sh --prefix PATH : $out/bin";
          };
        };
      }
    );
}
