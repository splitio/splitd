# ----- Builder image
FROM golang:1.20.7-alpine3.18 AS builder

RUN apk add git build-base bash

WORKDIR /splitd
COPY . .
RUN make clean splitd

# ----- Runner image
FROM alpine:3.18 AS runner

RUN apk add gettext yq bash
RUN mkdir -p /opt/splitd
COPY --from=builder /splitd/splitd /opt/splitd
COPY splitd.yaml.tpl /opt/splitd
COPY infra/entrypoint.sh /opt/splitd
RUN chmod +x /opt/splitd/entrypoint.sh

ENTRYPOINT ["/opt/splitd/entrypoint.sh"]
