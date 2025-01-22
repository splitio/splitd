# ----- Builder image
ARG GOLANG_VERSION=1.21.13
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

# ----- Runner image
FROM debian:bookworm-20250113-slim AS runner

ARG YQ_VERSION=v4.44.6

RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  bash ca-certificates wget socat && \
  wget -O /usr/local/bin/yq \
  "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" && \
  chmod +x /usr/local/bin/yq && \
  mkdir -p /opt/splitd && \
  rm -rf /var/lib/apt/lists/*

COPY --from=builder /splitd/splitd /opt/splitd
COPY --from=builder /splitd/splitd.yaml.tpl /opt/splitd
COPY infra/entrypoint.sh /opt/splitd
RUN chmod +x /opt/splitd/entrypoint.sh

ENTRYPOINT ["/opt/splitd/entrypoint.sh"]
