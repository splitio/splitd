# ----- Builder image
ARG GOLANG_VERSION=1.26.10
FROM golang:${GOLANG_VERSION}-trixie AS builder

ARG FIPS_MODE
ARG COMMIT_SHA
# yq is built from source (instead of downloading the release binary) so it is compiled
# for the target platform and with the same patched go toolchain used for splitd.
ARG YQ_VERSION=v4.54.1
# override yq's pinned x/net to pick up security fixes not yet released upstream
ARG YQ_XNET_VERSION=v0.61.0

RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  build-essential ca-certificates python3 git socat

RUN git clone --depth 1 --branch ${YQ_VERSION} https://github.com/mikefarah/yq.git /tmp/yq && \
  cd /tmp/yq && \
  go get golang.org/x/net@${YQ_XNET_VERSION} && \
  go mod tidy && \
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /usr/local/bin/yq . && \
  rm -rf /tmp/yq

WORKDIR /splitd
COPY . .

RUN export GITHUB_SHA="${COMMIT_SHA}" && bash -c '\
  if [[ "${FIPS_MODE}" = "enabled" ]]; \
  then echo "building in fips mode"; make clean splitd-fips splitd.yaml.tpl EXTRA_BUILD_ARGS="${EXTRA_BUILD_ARGS}"; mv splitd-fips splitd; \
  else echo "building in standard mode"; make clean splitd splitd.yaml.tpl EXTRA_BUILD_ARGS="${EXTRA_BUILD_ARGS}"; \
  fi'

# ----- Runner image
FROM debian:trixie-slim AS runner

RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get upgrade -y && \
  apt-get install --no-install-recommends -y \
  bash ca-certificates socat && \
  mkdir -p /opt/splitd && \
  rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/yq /usr/local/bin/yq
COPY --from=builder /splitd/splitd /opt/splitd
COPY --from=builder /splitd/splitd.yaml.tpl /opt/splitd
COPY infra/entrypoint.sh /opt/splitd
RUN chmod +x /opt/splitd/entrypoint.sh

ENTRYPOINT ["/opt/splitd/entrypoint.sh"]
