
antora_preview_version ?= 2.3.14
antora_preview_cmd ?= $(DOCKER_CMD) run --rm --publish 2020:2020 --volume "${PWD}":/preview/antora ghcr.io/vshn/antora-preview:latest --style=vshn

.PHONY: docs-preview
docs-preview: ## Preview the documentation
	$(antora_preview_cmd)
