# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-08-30

### Added

- feat: added configuration options for `go-test`. See [README.md](README.md#configuration-fields) for details. (#13)
- feat: added early exit on task failure (#14)
- feat: added `-C/--coverage-file` to provide a coverage file to do analysis on (#17)

### Changed

- feat: changed the flag `-t/--target-ref` to be required instead of optional (#18)

### Fixed

- fix: bug where cursor remains hidden when `ctrl+c` is pressed while tasks are running (#20)

## [0.2.0] - 2026-08-18

### Added

- feat: added flag `-o/--output` to write new code missing coverage out to a file (#8)
- feat: display time durations while running commands in background (#9)

### Changed

- refactor: changed header output to utilize bytes.Buffer for better performance (#4)
- refactor: missing coverage output to bold path and split each file by a newline for readability (#5)

## [0.1.1] - 2026-08-03

### Changed

- chore: Adjusted how versioning is set for builds to support both `.goreleaser.yml` and `go install`. (#2)
- chore: Added `-trimpath` to `.goreleaser.yml`. (#2)

## [0.1.0] - 2026-07-26

Initial development (#1)
