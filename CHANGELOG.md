# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2025-12-23

**Note:** Version 1.3.0 is the `latest` version, created to track the current development version.

### Added
- GitHub Actions workflow for automated tag and release management

## [1.2.0] - 2025-12-23

### Added
- Transfer webhook support
- CLI configuration management
- Local timezone support for cron tasks
- Enhanced metrics and monitoring

### Changed
- Improved cron task scheduling
- Enhanced CLI shell functionality
- Optimized event server and metrics

### Fixed
- Fixed panic issue in controller
- Fixed pod memory metrics
- Fixed event server and cron task issues
- Fixed installation and event resource handling

## [1.0.0] - 2024-11-07

### Added
- Initial release
- Task, TaskRun, Pipeline, PipelineRun CRDs
- Host and Cluster management
- EventHooks support
- Controller and Server components
- Helm chart support
- CLI tool (opscli)
- Web UI
- Prometheus and NATS integration

[1.3.0]: https://github.com/shaowenchen/ops/releases/tag/v1.3.0
[1.2.0]: https://github.com/shaowenchen/ops/releases/tag/v1.2.0
[1.0.0]: https://github.com/shaowenchen/ops/releases/tag/v1.0.0
