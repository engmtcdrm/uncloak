# Using uncloak in GitHub Actions

This repository provides a reusable GitHub Actions workflow at `.github/workflows/uncloak-action.yml`. Call it from a workflow in your repository to run `uncloak` across one or more operating systems and, for pull requests, post a coverage summary comment.

## Add a caller workflow

Create a workflow such as `.github/workflows/uncloak.yml` in your repository:

```yaml
name: Uncloak

on:
  pull_request:
    branches:
      - main

permissions:
  contents: read
  pull-requests: write

jobs:
  uncloak:
    uses: engmtcdrm/uncloak/.github/workflows/uncloak-action.yml@v0.7.0
    with:
      os_matrix: '["ubuntu-latest"]'
      go_version: "1.26.x"
      base_ref: ${{ github.base_ref }}
      head_ref: ${{ github.head_ref }}
      gate: true
```

Replace `main` with the branch you want to check, choose the Go version used by your project, and select an available [release ref](https://github.com/engmtcdrm/uncloak/releases) for the reusable workflow. You can also use a branch or commit SHA as the ref.

The workflow needs the `pull-requests: write` permission to create or update its pull request comment. If you do not want the caller workflow to grant that permission, remove it; the coverage check can still run, but the comment job will not have permission to post.

## Inputs

All inputs are required:

| Input | Type | Description |
| --- | --- | --- |
| `os_matrix` | string | JSON array of runner labels, for example `'["ubuntu-latest", "windows-latest"]'`. |
| `go_version` | string | Go version to install for the test run. |
| `base_ref` | string | Git reference to compare against, usually the pull request's base branch. |
| `head_ref` | string | Git reference containing the changes, usually the pull request's head branch. |
| `gate` | boolean | Whether a failing `uncloak` result should fail the workflow. |

For pull requests, `github.base_ref` and `github.head_ref` provide the base and head branch names. The current workflow checks out these refs by name, so make sure the head ref is available from the checked-out repository, especially when handling pull requests from forks.

## Coverage gate and results

With `gate: true`, the workflow fails when `uncloak` reports that coverage is below the configured threshold. With `gate: false`, the workflow still runs the check and uploads its summary, but a coverage failure alone does not fail the workflow.

When the caller is triggered by a `pull_request` event, the workflow creates or updates a comment with the coverage summary for each operating system in the matrix. It uses the repository's `.uncloak.yml` or `.uncloak.yaml` configuration when present; see the [configuration guide](../README.md#configuration) for details.

The reusable workflow installs the latest published version of the `uncloak` CLI on each run, regardless of the ref used to call the workflow.
