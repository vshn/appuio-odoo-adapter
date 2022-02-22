FROM docker.io/library/alpine:3.15 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    coreutils \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["appuio-odoo-adapter"]
COPY appuio-odoo-adapter /usr/bin/

COPY zone-names.yaml .
COPY description_templates/*.gotmpl description_templates/

USER 65536:0
