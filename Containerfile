ARG FEDORA_MAJOR_VERSION="${FEDORA_MAJOR_VERSION:-39}"

FROM registry.fedoraproject.org/fedora:${FEDORA_MAJOR_VERSION} AS builder

ARG APP_NAME="bext"
ENV OUTPUT_PATH=/app/output

WORKDIR /app 

ADD . /app

RUN dnf install \
    --disablerepo='*' \
    --enablerepo='fedora,updates' \
    --setopt install_weak_deps=0 \
    --nodocs \
    --assumeyes \
    'dnf-command(builddep)' \
    rpkg \
    rpm-build && \
    mkdir -p "$OUTPUT_PATH" && \
    rpkg spec --outdir  "$OUTPUT_PATH" && \
    dnf builddep -y output/$APP_NAME.spec && \
    rpkg local --outdir $PWD/output

FROM scratch

ENV OUTPUT_PATH=/app/output

COPY --from=builder ${OUTPUT_PATH} /out
