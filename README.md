# uncloak

`uncloak` is a CLI tool for analyzing new Go code coverage for the current working tree with a specified target reference that resolves to a commit (e.g. a branch name or commit hash).

At a high level, it:

- runs `git diff` against the resolved target commit and current working tree to find new Go lines
- runs `go test` to collect coverage data from the current working tree
- compares new Go lines from the diff against the coverage profile
- reports uncovered new lines and fails when coverage drops below the configurable threshold

## Installation

Install the latest version with:

```bash
go install github.com/engmtcdrm/uncloak@latest
```

This installs the `uncloak` binary into your Go `bin` directory.

## Usage

Run `uncloak` from the root of a Git repository:

```bash
uncloak --target-ref main
```

`uncloak` analyzes the current working tree against the target reference a user provides with `-t` / `--target-ref`. If the target ref resolves to the same commit as the current working tree, the diff analysis is skipped with a warning instead of failing.

If coverage is below the threshold, the command exits with an error. Regardless of the coverage result, it prints the uncovered new line ranges, if any.

## Configuration

`uncloak` reads configuration from a YAML file in the current working directory.

Supported file names:

- `.uncloak.yml`
- `.uncloak.yaml`

If no config file is present, `uncloak` uses built-in defaults. Empty config files are rejected.

### Default configuration

```yaml
version: 0
coverage-threshold: 80
go-test:
  count: 0
  timeout: 10m
  verbose: false
exclusions: []
```

### Configuration fields

- `version`: Config file version, currently only `0` is supported.
- `coverage-threshold`: Minimum acceptable coverage percentage for new code.
- `go-test`: Optional configuration for `go test` command, e.g. `-count`, `-timeout`, `-v`, etc.
  - `count`: Number of times to run tests. If this is set above 0 it will also ignore caching that `go test` does by default.
  - `timeout`: Test timeout duration, e.g. `30s`, `1m`, `2h`, etc. If not set, it will default to the `go test` default of 10 minutes.
  - `verbose`: Verbose output for `go test` command.
- `exclusions`: List of file paths or glob patterns to exclude from analysis.

### Example configuration

```yaml
version: 0
coverage-threshold: 90
go-test:
  count: 1
  timeout: 30s
  verbose: true
exclusions:
  - "docs/**"
  - "**/*_generated.go"
```

### Exclusions

Exclusions support exact file matches and glob patterns. For example:

- `main.go`
- `internal/**`
- `**/*_generated.go`

## Flags

`uncloak` supports these command-line flags:

- `-C, --coverage-file <string>`: (optional) path to the Go coverage file. If not specified, the default is to use the go tool to generate the coverage file
- `-c, --coverage-threshold <float>`: (optional) coverage threshold override. This will also overwrite what is specified in the configuration file
- `-d, --debug`: (optional) enable debug output, e.g. what commands are run
- `-h, --help`: help for uncloak
- `-o, --output`: (optional) file to write new code missing coverage out to
- `-t, --target-ref <string>`: (required) git target reference to compare against, e.g. a branch name or commit hash
- `-v, --verbose`: (optional) enable verbose output, e.g. output from go test command. This does not enable verbose go test. Use configuration file to enable verbose go test output
- `--version`: version for uncloak

Example:

```bash
uncloak --target-ref main --coverage-threshold 70.31 --verbose
```

## Exit status

- `0` when coverage meets the configured threshold
- non-zero when coverage is below the threshold or an analysis error occurs

## Example workflow

1. Create or switch to a feature branch.
2. Make changes and commit.
3. Run `uncloak` from the repository root.
4. Review any uncovered new lines.
5. Add tests or adjust code until the new coverage meets the threshold.

## Notes

- New branches must have one commit prior to being able to run `uncloak` against a target reference.
- Brand new Go files must be staged or committed for `uncloak` to analyze them; untracked files will not be part of the `git diff` analysis.
- The tool expects to run inside a Git repository.
