# ----- Builder image
ARG GOLANG_VERSION=1.26.1
FROM golang:${GOLANG_VERSION}-bookworm AS builder

ARG FIPS_MODE
ARG COMMIT_SHA

RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  build-essential ca-certificates python3 git socat

WORKDIR /splitd
COPY . .

RUN export GITHUB_SHA="${COMMIT_SHA}" && bash -c '\
  if [[ "${FIPS_MODE}" = "enabled" ]]; \
  then echo "building in fips mode"; make clean splitd-fips splitd.yaml.tpl EXTRA_BUILD_ARGS="${EXTRA_BUILD_ARGS}"; mv splitd-fips splitd; \
  else echo "building in standard mode"; make clean splitd splitd.yaml.tpl EXTRA_BUILD_ARGS="${EXTRA_BUILD_ARGS}"; \
  fi'

# Build yq from source with updated dependencies to avoid vulnerabilities
ARG YQ_VERSION=v4.52.4
WORKDIR /tmp/yq-build
RUN git clone --depth 1 --branch ${YQ_VERSION} https://github.com/mikefarah/yq.git . && \
    go get golang.org/x/net@v0.51.0 && \
    go mod tidy && \
    go build -o /go/bin/yq . && \
    cd /splitd && rm -rf /tmp/yq-build

# ----- Runner image
FROM debian:bookworm-slim AS runner

RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  bash ca-certificates socat && \
  mkdir -p /opt/splitd && \
  rm -rf /var/lib/apt/lists/*

COPY --from=builder /splitd/splitd /opt/splitd
COPY --from=builder /splitd/splitd.yaml.tpl /opt/splitd
COPY --from=builder /go/bin/yq /usr/local/bin/yq
COPY infra/entrypoint.sh /opt/splitd
RUN chmod +x /opt/splitd/entrypoint.sh

ENTRYPOINT ["/opt/splitd/entrypoint.sh"]