# ----- Builder image
FROM golang:1.21.6-bookworm AS builder

ARG FIPS_MODE
ARG COMMIT_SHA

RUN apt update -y
RUN apt install -y build-essential ca-certificates python3 git

WORKDIR /splitd
COPY . .

RUN export GITHUB_SHA="${COMMIT_SHA}" && \
    if [[ "${FIPS_MODE}" = "enabled" ]]; \
    then echo "building in fips mode"; make clean splitd-fips splitd.yaml.tpl EXTRA_BUILD_ARGS="${EXTRA_BUILD_ARGS}"; mv split-sync-fips split-sync; \
    else echo "building in standard mode"; make clean splitd splitd.yaml.tpl EXTRA_BUILD_ARGS="${EXTRA_BUILD_ARGS}"; \
    fi

# ----- Runner image

FROM debian:12.4 AS runner

RUN apt update -y
RUN apt install -y bash ca-certificates wget

RUN wget https://github.com/mikefarah/yq/releases/download/v4.40.5/yq_linux_amd64
RUN chmod +x yq_linux_amd64
RUN mv yq_linux_amd64 /usr/local/bin/yq

RUN mkdir -p /opt/splitd
COPY --from=builder /splitd/splitd /opt/splitd
COPY --from=builder /splitd/splitd.yaml.tpl /opt/splitd
COPY infra/entrypoint.sh /opt/splitd
RUN chmod +x /opt/splitd/entrypoint.sh

ENTRYPOINT ["/opt/splitd/entrypoint.sh"]
