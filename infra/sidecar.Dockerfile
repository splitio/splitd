# ----- Builder image
FROM golang:1.21.6-alpine3.19 AS builder

RUN apk add git build-base bash

WORKDIR /splitd
COPY . .
RUN make clean splitd splitd.yaml.tpl

# ----- Runner image
FROM alpine:3.19.0 AS runner

RUN apk add gettext yq bash
RUN mkdir -p /opt/splitd
COPY --from=builder /splitd/splitd /opt/splitd
COPY --from=builder /splitd/splitd.yaml.tpl /opt/splitd
COPY infra/entrypoint.sh /opt/splitd
RUN chmod +x /opt/splitd/entrypoint.sh

ENTRYPOINT ["/opt/splitd/entrypoint.sh"]
