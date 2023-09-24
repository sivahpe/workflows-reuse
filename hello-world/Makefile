# Copyright 2022 Hewlett Packard Enterprise Development LP

DOCKER := docker

VERSION = $(shell (git describe --long --tags --match 'v[0-9]*' || echo v0.0.0) | cut -c2-)
COMMIT  = $(shell git rev-parse --short HEAD)

DOCKER_IMAGE := hello-world
DOCKER_TAG := $(VERSION)

PUSH_IMAGE := $(DOCKER_IMAGE)
PUSH_TAG := $(DOCKER_TAG)

HTTPS_PROXY := ${HTTPS_PROXY}
HTTP_PROXY := ${HTTP_PROXY}

HELM_DIR = ./helm

LABEL_CREATED   = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LABEL_AUTHORS  := andrew.smith@hpe.com
LABEL_SOURCE   := https://github.hpe.com/cloud/ci_tools
LABEL_VERSION   = $(VERSION)
LABEL_REVISION  = $(COMMIT)
LABEL_VENDOR   := 'Hewlett Package Enterprise'
LABEL_TITLE    := $(DOCKER_IMAGE)

GO := go
GOBUILD := $(GO) build
LDFLAGS = -X main.Version=$(VERSION)

.PHONY: build
build:
	@mkdir -p dist
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o dist/ ./hello-world

.PHONY: golint
golint:
	golangci-lint --timeout 3m run

.PHONY: helmlint
helmlint:
	helm lint $(shell find $(HELM_DIR) -mindepth 1 -maxdepth 1 -type d)

.PHONY: lint
lint: golint helmlint

.PHONY: docker-build
docker-build:
	$(DOCKER) build \
		--pull \
		--force-rm \
		--target hello-world \
		--network host \
		--label org.opencontainers.image.created=$(LABEL_CREATED) \
		--label org.opencontainers.image.authors=$(LABEL_AUTHORS) \
		--label org.opencontainers.image.source=$(LABEL_SOURCE) \
		--label org.opencontainers.image.version=$(LABEL_VERSION) \
		--label org.opencontainers.image.revision=$(LABEL_REVISION) \
		--label org.opencontainers.image.vendor=$(LABEL_VENDOR) \
		--label org.opencontainers.image.title=$(LABEL_TITLE) \
		--build-arg HTTPS_PROXY=$(HTTPS_PROXY) \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		.
		@$(call built,$(DOCKER_IMAGE):$(DOCKER_TAG))

.PHONY: docker-build-multiple-targets
docker-build-multiple-targets:
	$(DOCKER) buildx build \
		--platform linux/amd64,linux/arm64 \
		--pull \
		--push \
		--force-rm \
		--target hello-world \
		--network host \
		--label org.opencontainers.image.created=$(LABEL_CREATED) \
		--label org.opencontainers.image.authors=$(LABEL_AUTHORS) \
		--label org.opencontainers.image.source=$(LABEL_SOURCE) \
		--label org.opencontainers.image.version=$(LABEL_VERSION) \
		--label org.opencontainers.image.revision=$(LABEL_REVISION) \
		--label org.opencontainers.image.vendor=$(LABEL_VENDOR) \
		--label org.opencontainers.image.title=$(LABEL_TITLE) \
		--build-arg HTTPS_PROXY=$(HTTPS_PROXY) \
		--build-arg VERSION=$(VERSION) \
		.
		@$(call built,$(DOCKER_IMAGE):$(DOCKER_TAG))

.PHONY: docker-push
docker-push:
	$(DOCKER) tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(PUSH_REGISTRY)/$(PUSH_PROJECT)/$(PUSH_IMAGE):$(PUSH_TAG)
	$(DOCKER) push $(PUSH_REGISTRY)/$(PUSH_PROJECT)/$(DOCKER_IMAGE):$(PUSH_TAG)

.PHONY: kubescore
kubescore:
	helm template helm/hello-world >> rendered.yaml
	kube-score score rendered.yaml --ignore-test pod-networkpolicy,networkpolicy-targets-pod,container-ephemeral-storage-request-and-limit
