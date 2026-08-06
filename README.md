# KAI Resource Management API

> [!WARNING]
> 🚧 **Work in progress** 🚧
>
> This module is still under active development 🛠️.
>
> APIs may change without notice, and it is not yet recommended for direct production use ⚠️.
>
> For now, consume it through
> [KAI Resource Management](https://github.com/kai-scheduler/kai-resource-management)
> unless you are prepared to track breaking changes closely.

Kubernetes API types and generated CRD manifests for the `kai.resources` API group, used
by [KAI Resource Management](https://github.com/kai-scheduler/kai-resource-management) —
the organizational layer around
[KAI Scheduler](https://github.com/NVIDIA/KAI-Scheduler) covering projects, departments,
node pools, and queues.

This is a standalone Go module so the API can be versioned and consumed independently of
the controllers that implement it, the same relationship
[`github.com/kai-scheduler/api`](https://github.com/kai-scheduler/api) has with KAI
Scheduler.

**This repository contains only the API contract:** CRD type definitions, their generated
DeepCopy implementations, and the generated CRD manifests. Controllers, the Helm chart,
installation instructions, and user documentation all live in
[kai-resource-management](https://github.com/kai-scheduler/kai-resource-management).

The API types are being introduced incrementally. This release contains the
`kai.resources/v1alpha1` group registration; the resource types — **Project**,
**Department**, **NodePool**, and **ManagedNodesConfig** — follow.

## Installation

```bash
go get github.com/kai-scheduler/kai-resource-management-api
```

## Usage

Register the types with a scheme and use `controller-runtime`'s client. Generated
clientsets, informers, and listers are deliberately not produced.

```go
import (
    kaires "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
    "k8s.io/apimachinery/pkg/runtime"
)

scheme := runtime.NewScheme()
if err := kaires.AddToScheme(scheme); err != nil {
    return err
}
```

## CRD manifests

Generated manifests live in `config/crd/`. `kai-resource-management` syncs them from the
pinned module version into its Helm chart, so they are a released artifact rather than a
build byproduct.

## Development

Requires the Go version in `go.mod`, Make, and Git. No Docker, Helm, or Kubernetes
cluster.

```bash
make build      # go build ./...
make test       # go test ./...
make lint       # gofmt -w + go vet
make validate   # gofmt check + go mod tidy -diff + go vet   (what CI runs)
```

DeepCopy implementations and CRD manifests are generated with `controller-gen`, pinned by
the Makefile and installed into the ignored `bin/` directory:

```bash
make generate   # DeepCopy implementations
make manifests  # CRD manifests into config/crd/
make all        # both
```

Run `make all` after changing API types and commit the result in the same pull request. A
non-empty `git diff` afterwards means the committed generated output does not match the
Go types — and because `kai-resource-management` consumes `config/crd/` directly, that
mismatch ships to users.

Keep the `controller-gen` version aligned with `kai-resource-management`, so a CRD
manifest does not depend on which repository generated it.

## Versioning and releases

Follows [Semantic Versioning](https://semver.org/); `v0.x` until the contract is declared
stable. A release is a `vX.Y.Z` tag on `main` — the Go module proxy serves it from there.
Consumers bump with:

```bash
go get github.com/kai-scheduler/kai-resource-management-api@vX.Y.Z
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Conventions for both human and automated
contributors are in [AGENTS.md](AGENTS.md).

## License

Apache 2.0. Copyright NVIDIA CORPORATION.
