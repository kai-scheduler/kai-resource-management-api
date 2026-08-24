# KAI Resource Management API — Agent Development Guide

`github.com/kai-scheduler/kai-resource-management-api` is a standalone Go module holding
the Kubernetes API types and generated CRD manifests for the `kai.resources` API group.
It is the client-facing contract that
[kai-resource-management](https://github.com/kai-scheduler/kai-resource-management)
implements and consumes as a versioned dependency, mirroring the relationship between
[KAI Scheduler](https://github.com/NVIDIA/KAI-Scheduler) and
[`github.com/kai-scheduler/api`](https://github.com/kai-scheduler/api).

**This repository is deliberately minimal.** It contains CRD types, their generated
DeepCopy implementations, and the generated manifests — nothing else. Before adding any
file, justify it against that. If the justification depends on something this repository
does not have, leave it out.

## Commands

```bash
make build      # go build ./...
make test       # go test ./...
make lint       # gofmt -w + go vet
make validate   # gofmt check + go mod tidy -diff + go vet   (what CI runs)
make generate   # DeepCopy implementations
make manifests  # CRD manifests into config/crd/
make all        # generate + manifests
```

## What belongs here

**Include:** CRD type definitions for the `kai.resources` group — including `KRMConfig`,
the cluster-scoped singleton the KRM operator reconciles — generated DeepCopy
implementations, generated CRD manifests, and public API constants or small helpers for
working with the types.

**Exclude:** controllers, webhooks, the Helm chart, deployment configuration, and feature
flags. Those live in `kai-resource-management`.

A served type is part of the API contract even when its fields configure runtime
behaviour, so `KRMConfig` belongs here: users read and write it with `kubectl`, and its
schema is versioned and released like any other kind. The configuration that stays out is
the kind that is never served — chart `values.yaml`, chart templates, and operator
feature flags.

This module may depend on sibling API modules such as `github.com/kai-scheduler/api`,
but never on a service repository.

Do not add: nested `go.mod` or `go.work` files; `cmd/`, `pkg/`, `deployments/`, or
`build/` directories; a top-level `examples` directory; or generated clientsets,
informers, and listers — consumers use `controller-runtime`'s client with `AddToScheme`.

## Layout

- `kai/<version>/` — API types, one package per API group version.
- `config/crd/` — generated CRD manifests. Never hand-edited.
- `hack/` — copyright boilerplate for the code generators.

## Code generation

Run `make all` after changing API types and commit the result in the same pull request.
`kai-resource-management` consumes `config/crd/` directly from the released module, so
generated output that is out of date with the Go types ships broken CRDs to users. Treat
a dirty `git diff` after `make all` as a build failure.

Keep the `controller-gen` version aligned with `kai-resource-management`.

## Conventions

- Go files use `snake_case.go`; type files end in `_types.go`.
- Every field needs a JSON tag; optional fields need `omitempty` and `// +optional`.
- Put kubebuilder markers immediately above the declaration they affect.
- Additive changes only within a released version. A breaking change to a served type
  needs a new version, not an edit.
- Field GoDoc becomes the CRD description users see from `kubectl explain`. Write it for
  them.
- Apache-2.0 + NVIDIA SPDX headers on every hand-written file.
- Import groups: standard library, external, then this module.

## Testing

Test what generation cannot guarantee: CEL validation rules, defaulting, conversion, and
hand-written helpers. Do not assert on generated DeepCopy implementations. A test that
reads a generated CRD manifest and asserts on it is valuable — it is the only thing tying
the Go types to the shipped YAML.

`make test` must stay safe on a developer machine and must not connect to or mutate a
Kubernetes cluster.

## Pull requests

Titles follow Conventional Commits (`feat`, `fix`, `docs`, `refactor`, `test`, `build`,
`ci`, `chore`). Every commit must be signed off — see `CLA.md`.

Record user-visible changes under `## [Unreleased]` in `CHANGELOG.md` directly; there is
no fragment tooling here.

Before finishing: run `make all` and confirm the tree is clean, then `make validate`.
