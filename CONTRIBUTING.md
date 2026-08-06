# Contributing to the KAI Resource Management API

Thank you for contributing.

This repository holds only the Kubernetes API types and generated CRD manifests for the
`kai.resources` API group. The controllers that implement them, the Helm chart, and the
user documentation live in
[kai-resource-management](https://github.com/kai-scheduler/kai-resource-management),
which consumes this module as a versioned dependency. API changes land here first; the
consumer then bumps the dependency.

Read the [Code of Conduct](CODE_OF_CONDUCT.md) and the
[Developer Certificate of Origin](CLA.md) before submitting a change.

## Getting started

1. Fork and clone the repository.
2. Create a focused branch from `main`.
3. Make the change, including tests where the behavior is not generated.
4. Run `make all` and commit any regenerated output.
5. Run `make validate`.
6. Commit with DCO sign-off and open a pull request.

Repository conventions for both human and automated contributors are in
[AGENTS.md](AGENTS.md).

## Code generation

DeepCopy implementations and CRD manifests are generated. After changing API types:

```bash
make all
git diff --exit-code
```

A non-empty diff means the committed generated output does not match the Go types.
`kai-resource-management` consumes `config/crd/` directly from the released module, so
that mismatch ships to users.

## Pull requests

Titles follow [Conventional Commits](https://www.conventionalcommits.org/), for example
`feat(project): add deletionType to ProjectSpec`. Mark breaking changes with `!`.

Every commit must be signed off to certify the
[Developer Certificate of Origin](CLA.md):

```bash
git commit -s -m "feat(project): add deletionType to ProjectSpec"
```

Record user-visible changes under `## [Unreleased]` in
[CHANGELOG.md](CHANGELOG.md). There is no fragment tooling in this repository — edit the
file directly.

## Releasing

Maintainers release by tagging `vX.Y.Z` on `main`; the Go module proxy serves it from
there. Move the `## [Unreleased]` entries under the new version heading in the same
commit.

## Reporting issues

Bugs and feature requests about resource management behavior belong in the
[kai-resource-management issue tracker](https://github.com/kai-scheduler/kai-resource-management/issues).
Open an issue here only for the API types themselves.

Security vulnerabilities must be reported privately — see [SECURITY.md](SECURITY.md).
