# uncloak

`uncloak` is a CLI tool for analyzing new Go code coverage on the current branch with a target ref (by default `origin/main`).

At a high level, it:

- runs `git diff` against the configured target ref
- runs Go tests to collect coverage data
- compares added Go lines against the coverage profile
- reports uncovered new lines and fails when coverage drops below the configured threshold

## Installation

Install the latest version with:

```bash
go install github.com/engmtcdrm/uncloak@latest
```

This installs the `uncloak` binary into your Go `bin` directory.

## Usage

Run `uncloak` from the root of a Git repository:

```bash
uncloak
```

By default, `uncloak` will analyze against `origin/main`.

If coverage is below the threshold, the command exits with an error and prints the uncovered new line ranges.

## Configuration

`uncloak` reads configuration from a YAML file in the current working directory.

Supported file names:

- `.uncloak.yml`
- `.uncloak.yaml`

If no config file is present, `uncloak` uses built-in defaults.

### Default configuration

```yaml
version: 0
coverage-threshold: 80
exclusions: []
git:
  unstaged: true
  target-ref: origin/main
```

### Configuration fields

- `version`: config file version
- `coverage-threshold`: minimum acceptable coverage percentage for new code
- `exclusions`: list of file paths or glob patterns to exclude from analysis
- `git.unstaged`: include unstaged changes in the diff
- `git.target-ref`: target ref to compare against, such as `main` or `origin/main`

### Example configuration

```yaml
version: 0
coverage-threshold: 90
exclusions:
  - "docs/**"
  - "**/*_generated.go"
git:
  unstaged: true
  target-ref: main
```

### Exclusions

Exclusions support exact file matches and glob patterns. For example:

- `main.go`
- `internal/**`
- `**/*_generated.go`

## Flags

`uncloak` supports these command-line flags:

- `-t, --coverage-threshold <float>`: override the minimum coverage threshold. This will also overwrite what is specified in the configuration file.
- `-v, --verbose`: print the raw Go test output

Example:

```bash
uncloak --coverage-threshold 70.31 --verbose
```

## Exit status

- `0` when coverage meets the configured threshold
- non-zero when coverage is below the threshold or an analysis error occurs

## Example workflow

1. Create or switch to a feature branch.
2. Make changes.
3. Run `uncloak` from the repository root.
4. Review any uncovered new lines.
5. Add tests or adjust code until the new coverage meets the threshold.

## Notes

- The tool expects to run inside a Git repository.
- The default target ref is `origin/main`.
- The default coverage threshold is `80%`.
- Unknown YAML fields are rejected, so config files should only contain supported keys.
