FROM docker.io/library/alpine:3.15 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["odoo-adapter"]
COPY odoo-adapter /usr/bin/

USER 65536:0
