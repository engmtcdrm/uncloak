# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- feat: added configuration options for `go-test`. See [README.md](README.md#configuration-fields) for details.

## [0.2.0] - 2026-08-18

### Added

- feat: added flag `-o/--output` to write new code missing coverage out to a file
- feat: display time durations while running commands in background

### Changed

- refactor: changed header output to utilize bytes.Buffer for better performance
- refactor: missing coverage output to bold path and split each file by a newline for readability

## [0.1.1] - 2026-08-03

### Changed

- chore: Adjusted how versioning is set for builds to support both `.goreleaser.yml` and `go install`.
- chore: Added `-trimpath` to `.goreleaser.yml`.

## [0.1.0] - 2026-07-26

Initial development
