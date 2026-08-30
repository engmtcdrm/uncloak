package gitdiff

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const hunkHeaderPattern = `@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`

var hunkHeaderRegex = regexp.MustCompile(hunkHeaderPattern)

type parser struct {
	RawGitDiffOutput []byte
	Command          string
}

// Run executes the 'git diff' command with the provided options and parses its
// output into a [Results] struct. If opts is nil, it uses [DefaultOptions].
func Run(ctx context.Context, opts *Options) (*Results, error) {
	if opts == nil {
		opts = &DefaultOptions
	}

	if !isGitDir(ctx) {
		return nil, ErrNotAGitRepo
	}

	currentBranch := getCurrentBranch(ctx)
	if currentBranch == "" {
		return nil, ErrNoCurrentBranch
	}

	if opts.TargetRef == currentBranch {
		return nil, NewSameBranchError(opts.TargetRef, currentBranch)
	}

	p := &parser{}

	return p.runAndParseGitDiff(ctx, opts)
}

// parseHunkHeader checks if the line is a hunk header and if so, it updates the
// plusStartLine to the new start line of the hunk. It returns the updated
// plusStartLine and a boolean indicating whether the line was a hunk header or
// not.
func parseHunkHeader(line string, plusStartLine int) (int, bool) {
	if !hunkHeaderRegex.MatchString(line) {
		return plusStartLine, false
	}

	matches := hunkHeaderRegex.FindStringSubmatch(line)
	if len(matches) >= 3 {
		newStart, _ := strconv.Atoi(matches[2])
		plusStartLine = newStart
	}

	return plusStartLine, true
}

// parseGitDiffData reads the git diff data from the provided reader and parses
// it into a [Results] struct.
func (p *parser) parseGitDiffData(ctx context.Context, r io.Reader) (*Results, error) {
	s := bufio.NewScanner(r)
	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, nil
	}

	return p.parseLines(ctx, lines)
}

// parseLines processes the lines of a git diff and extracts the added lines for
// Go files and returns them in a [Results] struct.
func (p *parser) parseLines(ctx context.Context, lines []string) (*Results, error) {
	results, err := NewResults(ctx, p.Command)
	if err != nil {
		return nil, err
	}

	var currentFile string
	var plusStartLine int

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++ "):
			idx := strings.Index(line, "/")
			path := line[idx+1:]
			currentFile = path
			continue
		case strings.HasPrefix(line, "-"):
			continue
		case currentFile == "" || !isGoFile(currentFile):
			continue
		}
		var isHunkHeader bool
		if plusStartLine, isHunkHeader = parseHunkHeader(line, plusStartLine); isHunkHeader {
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++ ") {
			diffMap, ok := results.NewLines[currentFile]
			if !ok {
				diffMap = make(map[int]bool)
				results.NewLines[currentFile] = diffMap
			}

			diffMap[plusStartLine] = true
		}

		plusStartLine++
	}

	return results, nil
}

// runAndParseGitDiff executes the git diff command with the provided options,
// captures its output, and parses it into a [Results] struct.
func (p *parser) runAndParseGitDiff(ctx context.Context, opts *Options) (*Results, error) {
	cmd := exec.CommandContext(ctx, "git", "diff")

	args := optionsToArgs(opts)

	cmd.Args = append(cmd.Args, args...)

	p.Command = strings.Join(cmd.Args, " ")

	output, err := cmd.Output()
	if err != nil {
		return nil, handleExecError(cmd, output, err)
	}

	if len(output) == 0 {
		return nil, errNoOutput(ctx, cmd, opts.TargetRef)
	}

	return p.parseGitDiffData(ctx, bytes.NewReader(output))
}
