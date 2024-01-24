FROM bootstrap-nix as builder
WORKDIR /tmp/build

RUN \
  --mount=type=cache,target=/nix,from=nixos/nix:latest,source=/nix \
  --mount=type=cache,target=/root/.cache \
  --mount=type=bind,target=/tmp/build \
  <<EOF
    nix build . --out-link /tmp/output/result \
    nix copy /tmp/output/result --to /nix-store-closure
  EOF

FROM scratch
WORDKDIR /storage
COPY --from=builder /nix-store-closure /nix/store
COPY
