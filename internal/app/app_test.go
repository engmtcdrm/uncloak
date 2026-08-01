package app

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for [setVersion] function.
func Test_setVersion(t *testing.T) {
	t.Run("should not set version if already set", func(t *testing.T) {
		v := "1.0.0"
		setVersion(&v)
		require.Equal(t, "1.0.0", v, "Expected version to remain '1.0.0'")
	})

	t.Run("should leave version empty when build info is unavailable", func(t *testing.T) {
		originalReadBuildInfo := readBuildInfo
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return nil, false
		}
		defer func() {
			readBuildInfo = originalReadBuildInfo
		}()

		v := ""
		setVersion(&v)
		require.Empty(t, v)
	})

	t.Run("should set version from build info when available", func(t *testing.T) {
		expectedVersion := "v0.1.0"
		originalReadBuildInfo := readBuildInfo
		readBuildInfo = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: expectedVersion}}, true
		}
		defer func() {
			readBuildInfo = originalReadBuildInfo
		}()

		v := ""
		setVersion(&v)
		require.Equal(t, expectedVersion, v)
	})
}
