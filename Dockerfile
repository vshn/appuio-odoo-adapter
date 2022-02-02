FROM docker.io/library/alpine:3.15 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["appuio-odoo-adapter"]
COPY appuio-odoo-adapter /usr/bin/

USER 65536:0
