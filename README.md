# uncloak

`uncloak` is a CLI tool for analyzing new Go code coverage on the current branch with the current branch's nearest parent branch.

At a high level, it:

- runs `git diff` against the parent branch
- runs Go tests to collect coverage data
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
uncloak
```

By default, `uncloak` will analyze against the nearest parent branch of the current branch.
That default behavior requires a branch with a parent branch, so feature branches are the intended use.

If coverage is below the threshold, the command exits with an error and prints the uncovered new line ranges.

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
exclusions: []
```

### Configuration fields

- `version`: config file version
- `coverage-threshold`: minimum acceptable coverage percentage for new code
- `exclusions`: list of file paths or glob patterns to exclude from analysis

### Example configuration

```yaml
version: 0
coverage-threshold: 90
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

- `-c, --coverage-threshold <float>`: (optional) coverage threshold override. This will also overwrite what is specified in the configuration file
- `-d, --debug`: enable debug output, e.g. what commands are run
- `-t, --target-ref <string>`: git target ref to compare against
- `-v, --verbose`: enable verbose output, e.g. output from go test command

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

- Brand new Go files must be staged or committed for `uncloak` to analyze them.
- The tool expects to run inside a Git repository on a branch with a parent branch.
- The default coverage threshold is `80%`.
- Unknown YAML fields are rejected, so config files should only contain supported keys.
