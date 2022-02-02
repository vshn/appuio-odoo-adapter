## These are some common variables for Make

PROJECT_ROOT_DIR = .
PROJECT_NAME ?= appuio-odoo-adapter
PROJECT_OWNER ?= vshn

## BUILD:go
BIN_FILENAME ?= appuio-odoo-adapter

## BUILD:docker
DOCKER_CMD ?= docker

IMG_TAG ?= latest
# Image URL to use all building/pushing image targets
CONTAINER_IMG ?= local.dev/$(PROJECT_OWNER)/$(PROJECT_NAME):$(IMG_TAG)
