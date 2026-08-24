# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Record user-visible changes under `## [Unreleased]` as part of the pull request that makes
them. At release time, move those entries under a new `## [vX.Y.Z] - YYYY-MM-DD` heading.

## [Unreleased]

## [v0.1.1] - 2026-08-24

### Added

- `KRMConfig` — the cluster-scoped singleton that configures the KRM operator and the
  services it deploys — together with the `NodePoolController`, `ProjectController` and
  `PodGroupAssigner` types nested in its spec, and the generated
  `config/crd/kai.resources_krmconfigs.yaml` manifest.
- `kai/v1alpha1/schedulingshardargs` — the KAI Scheduler CLI flag names used as keys in
  `SchedulingShardConfig.Args`, mirrored here because KAI Scheduler does not export them.

## [v0.1.0] - 2026-08-06

### Added

- Initial release of the `kai.resources/v1alpha1` API types — `Project`, `Department`,
  `NodePool` and `ManagedNodesConfig` — with generated DeepCopy implementations and CRD
  manifests.
