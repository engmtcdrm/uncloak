package gitdiff

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/engmtcdrm/uncloak/internal/testing/testgit"
	"github.com/engmtcdrm/uncloak/internal/testing/testrepo"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [Run] function.
func Test_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("should return results with valid options", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		opts := &DefaultOptions

		results, err := Run(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
	})

	t.Run("should return results when opts is nil", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		results, err := Run(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, results)
	})

	t.Run("should return error when current directory is not a git repository", func(t *testing.T) {
		t.Chdir(t.TempDir())
		opts := &DefaultOptions

		results, err := Run(ctx, opts)
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("should return error when there is no parent branch", func(t *testing.T) {
		tempDir, _ := testrepo.Init(ctx, t)
		t.Chdir(tempDir)
		opts := &DefaultOptions

		results, err := Run(ctx, opts)
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("should return error if git is in a detached HEAD state", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		cmd := exec.CommandContext(ctx, "git", "checkout", "--detach", "HEAD")
		require.NoError(t, cmd.Run())

		opts := &DefaultOptions

		results, err := Run(ctx, opts)
		require.Error(t, err)
		require.Nil(t, results)
	})
}

// Tests for [parseHunkHeader] function.
func Test_parseHunkHeader(t *testing.T) {
	t.Run("valid hunk header", func(t *testing.T) {
		line := "@@ -1,5 +1,6 @@"
		plusStartLine := 1

		newPlusStartLine, isHunkHeader := parseHunkHeader(line, plusStartLine)
		require.True(t, isHunkHeader)
		require.Equal(t, 1, newPlusStartLine)
	})

	t.Run("invalid hunk header", func(t *testing.T) {
		line := "This is not a hunk header"
		plusStartLine := 1

		newPlusStartLine, isHunkHeader := parseHunkHeader(line, plusStartLine)
		require.False(t, isHunkHeader)
		require.Equal(t, plusStartLine, newPlusStartLine)
	})
}

// Tests for [parseGitDiffData] function.
func Test_parseGitDiffData(t *testing.T) {
	ctx := context.Background()
	p := &parser{}
	diffData := `diff --git a/internal/gitdiff/parser.go b/internal/gitdiff/parser.go
new file mode 100644
index 0000000..6d36618
--- /dev/null
+++ b/internal/gitdiff/parser.go
@@ -0,0 +1,2 @@
+This is a new line
+This is another existing line`

	t.Run("Should return error if scanner encounters an error", func(t *testing.T) {
		r := &testutils.ErrorReader{}
		results, err := p.parseGitDiffData(ctx, r)
		require.Error(t, err)
		require.Nil(t, results)
	})

	t.Run("should return nil results for empty input", func(t *testing.T) {
		r := &testutils.EmptyReader{}
		results, err := p.parseGitDiffData(ctx, r)
		require.NoError(t, err)
		require.Nil(t, results)
	})

	t.Run("should parse valid git diff data", func(t *testing.T) {
		r := strings.NewReader(diffData)
		results, err := p.parseGitDiffData(ctx, r)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Len(t, results.Files(), 1)
		assert.Equal(t, map[int]bool{1: true, 2: true}, results.NewLines["internal/gitdiff/parser.go"])
	})

	t.Run("should return error if parseLines fails", func(t *testing.T) {
		t.Chdir(t.TempDir())

		r := strings.NewReader(diffData)
		results, err := p.parseGitDiffData(ctx, r)
		require.Error(t, err)
		require.Nil(t, results)
	})
}

// Tests for [parseLines] function.
func Test_parseLines(t *testing.T) {
	ctx := context.Background()
	p := &parser{}

	t.Run("should correctly identify new lines on new files", func(t *testing.T) {
		lines := []string{
			"diff --git a/internal/gitdiff/parser.go b/internal/gitdiff/parser.go",
			"new file mode 100644",
			"index 0000000..6d36618",
			"--- /dev/null",
			"+++ b/internal/gitdiff/parser.go",
			"@@ -0,0 +1,2 @@",
			"+This is a new line",
			"+This is another existing line",
			"diff --git a/internal/config/config.go b/internal/config/config.go",
			"new file mode 100644",
			"index 0000000..1111111",
			"--- /dev/null",
			"+++ b/internal/config/config.go",
			"@@ -0,0 +1,3 @@",
			"+type Config struct {}",
			"+func (c *Config) Validate() error { return nil }",
			"+func Load() (*Config, error) { return nil, nil }",
			"diff --git a/internal/analyzer/analyzer.go b/internal/analyzer/analyzer.go",
			"new file mode 100644",
			"index 0000000..2222222",
			"--- /dev/null",
			"+++ b/internal/analyzer/analyzer.go",
			"@@ -0,0 +1,4 @@",
			"+func NewCodeCoverage() {}",
			"+func analyzeCoverage() {}",
			"+func processFiles() {}",
			"+func report() {}",
		}

		results, err := p.parseLines(ctx, lines)
		require.NoError(t, err)
		assert.Len(t, results.Files(), 3)
		assert.Equal(t, map[int]bool{1: true, 2: true}, results.NewLines["internal/gitdiff/parser.go"])
		assert.Equal(t, map[int]bool{1: true, 2: true, 3: true}, results.NewLines["internal/config/config.go"])
		assert.Equal(t, map[int]bool{1: true, 2: true, 3: true, 4: true}, results.NewLines["internal/analyzer/analyzer.go"])
	})

	t.Run("should return error if NewResults fails", func(t *testing.T) {
		t.Chdir(t.TempDir())

		lines := []string{
			"diff --git a/internal/gitdiff/parser.go b/internal/gitdiff/parser.go",
			"new file mode 100644",
			"index 0000000..6d36618",
			"--- /dev/null",
			"+++ b/internal/gitdiff/parser.go",
			"@@ -0,0 +1,2 @@",
			"+This is a new line",
			"+This is another existing line",
		}

		results, err := p.parseLines(ctx, lines)
		require.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("should return correct new lines for cmd/root.go", func(t *testing.T) {
		lines := []string{
			"diff --git a/cmd/root.go b/cmd/root.go",
			"index 248d540..1e4006c 100644",
			"--- a/cmd/root.go",
			"+++ b/cmd/root.go",
			"@@ -6,6 +6,8 @@ import (",
			"",
			"        \"github.com/spf13/cobra\"",
			"",
			"+       \"github.com/engmtcdrm/go-entomb/crypt\"",
			"+",
			"        \"github.com/engmtcdrm/mellon/internal/app\"",
			"        \"github.com/engmtcdrm/mellon/internal/cli/createcmd\"",
			"        \"github.com/engmtcdrm/mellon/internal/cli/deletecmd\"",
			"@@ -35,24 +37,34 @@ func init() {",
			"        mkdir(env.Instance.SecretsPath(), constants.SecureDirMode)",
			"        secureFiles(env.Instance.AppHomeDir(), constants.SecureDirMode, constants.SecureFileMode)",
			"",
			"-       secretFiles, err := secrets.GetSecretFiles(",
			"-               env.Instance.KeyPath(),",
			"-               env.Instance.SecretsPath(),",
			"-               env.Instance.SecretExt(),",
			"-       )",
			"+       secretCrypt, err := crypt.NewCrypt(env.Instance.KeyPath(), env.Instance.SecretsPath(), false, true)",
			"        if err != nil {",
			"                fmt.Println(err)",
			"                os.Exit(1)",
			"        }",
			"+       secretCrypt.TombFileExt(env.Instance.SecretExt())",
			"+",
			"+       var secretFiles []secrets.Secret",
			"+       for _, tomb := range secretCrypt.Tombs() {",
			"+               newSecret, err := secrets.NewSecret(env.Instance.KeyPath(), tomb.Name(), tomb.Path())",
			"+               if err != nil {",
			"+                       fmt.Println(err)",
			"+                       os.Exit(1)",
			"+               }",
			"+",
			"+               secretFiles = append(secretFiles, *newSecret)",
			"+       }",
			"",
			"        rootCmd.SilenceUsage = true",
			"        rootCmd.CompletionOptions.DisableDefaultCmd = true",
			"",
			"-       rootCmd.AddCommand(createcmd.NewCommand(secretFiles))",
			"-       rootCmd.AddCommand(deletecmd.NewCommand(secretFiles))",
			"-       rootCmd.AddCommand(listcmd.NewCommand(secretFiles))",
			"-       rootCmd.AddCommand(updatecmd.NewCommand(secretFiles))",
			"-       rootCmd.AddCommand(viewcmd.NewCommand(secretFiles))",
			"+       rootCmd.AddCommand(",
			"+               createcmd.NewCommand(secretCrypt, secretFiles),",
			"+               deletecmd.NewCommand(secretFiles),",
			"+               listcmd.NewCommand(secretCrypt),",
			"+               updatecmd.NewCommand(secretFiles),",
			"+               viewcmd.NewCommand(secretCrypt),",
			"+       )",
			" }",
			"",
			" // Execute executes the root command.",
		}

		expectedLines := map[string]map[int]bool{
			"cmd/root.go": {
				9:  true,
				10: true,
				40: true,
				45: true,
				46: true,
				47: true,
				48: true,
				49: true,
				50: true,
				51: true,
				52: true,
				53: true,
				54: true,
				55: true,
				56: true,
				61: true,
				62: true,
				63: true,
				64: true,
				65: true,
				66: true,
				67: true,
			},
		}

		results, err := p.parseLines(ctx, lines)
		require.NoError(t, err)
		require.Equal(t, expectedLines, results.NewLines)
	})
}

// Tests for [runAndParseGitDiff] function.
func Test_runAndParseGitDiff(t *testing.T) {
	ctx := context.Background()
	p := &parser{}

	t.Run("should return error if current directory is not a git repository", func(t *testing.T) {
		t.Chdir(t.TempDir())
		opts := &DefaultOptions

		results, err := p.runAndParseGitDiff(ctx, opts)
		require.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("should return no output error for valid git diff command with no changes", func(t *testing.T) {
		opts := &Options{
			TargetRef: testgit.MainBranchName,
		}

		_, _ = testrepo.Init(ctx, t)

		results, err := p.runAndParseGitDiff(ctx, opts)
		require.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("should return results for valid git diff command", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		opts := &DefaultOptions

		results, err := p.runAndParseGitDiff(ctx, opts)
		require.NoError(t, err)
		assert.NotNil(t, results)
	})

	t.Run("should return results for valid git diff command with debug true", func(t *testing.T) {
		_, _ = testrepo.InitWithFileCopy(ctx, t)
		opts := &DefaultOptions

		results, err := p.runAndParseGitDiff(ctx, opts)
		require.NoError(t, err)
		assert.NotNil(t, results)
	})
}
