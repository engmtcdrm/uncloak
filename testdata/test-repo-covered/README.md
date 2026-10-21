# README.md

This is a test repository exists for testing the uncloak package. Uncloak logic that involves running Go tests cannot be run against uncloak itself or it will cause an infinite loop. Thus this repository exists for testing uncloak logic that involve running Go tests. Most of the files are typical files you'd find in a repository with the exception of the following:

- `magic_50_test.go` - Provides 50% test coverage of this test repository.
- `magic_100_test.go` - Provides 100% test coverage of this test repository.

These files can be removed to force coverage to fail or succeed.

This repository will need to utilize the `testgit` package to initialize a dummy git repository in a test directory, e.g. `t.TempDir()` during testing. See any of the tests in uncloak utilizing this or the `testgit` package to see how this can be done.
