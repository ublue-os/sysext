FROM nixos/nix

COPY . /bakery
RUN chown -R 1000:1000 /bakery
WORKDIR /bakery

RUN nix-channel --update
RUN nix-env -iA nixpkgs.squashfsTools
