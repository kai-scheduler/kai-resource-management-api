# Copyright 2026 NVIDIA CORPORATION
# SPDX-License-Identifier: Apache-2.0

# Keep this Makefile aligned with kai-resource-management, so a CRD manifest does not
# depend on which repository generated it.
CONTROLLER_TOOLS_VERSION ?= v0.20.1

GO ?= go
LOCALBIN ?= $(CURDIR)/bin
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

.PHONY: all
all: generate manifests ## Regenerate everything derived from the API types.

.PHONY: build
build: ## Build all packages.
	$(GO) build ./...

.PHONY: test
test: ## Run all tests.
	$(GO) test ./...

.PHONY: lint
lint: ## Format Go files and run go vet.
	gofmt -l -w .
	$(GO) vet ./...

.PHONY: validate
validate: ## Verify formatting, module tidiness, and vet without changing files.
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		printf 'Go files require formatting:\n%s\n' "$$unformatted"; \
		exit 1; \
	fi
	$(GO) mod tidy -diff
	$(GO) vet ./...

.PHONY: generate
generate: $(CONTROLLER_GEN) ## Generate DeepCopy implementations.
	$(CONTROLLER_GEN) object:headerFile="./hack/boilerplate.go.txt" paths="./kai/..."

# allowDangerousTypes permits the floating-point quota fields in the API types.
.PHONY: manifests
manifests: $(CONTROLLER_GEN) ## Generate CRD manifests into config/crd.
	$(CONTROLLER_GEN) crd:allowDangerousTypes=true,generateEmbeddedObjectMeta=true,headerFile="./hack/boilerplate.yaml.txt" paths="./kai/..." output:crd:artifacts:config=config/crd

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

$(CONTROLLER_GEN): | $(LOCALBIN)
	GOBIN=$(LOCALBIN) $(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
