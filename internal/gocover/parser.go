package gocover

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type parser struct {
	GoList          *GoList
	RawCoverData    []byte
	RawGoTestOutput []byte
}

// Run executes the 'go test' command to generate a coverage profile. Finally it
// parses the coverage profile and returns it as a [Profile] struct.
func Run() (*Profile, error) {
	goList, err := getGoList()
	if err != nil {
		return nil, err
	}

	p := &parser{
		GoList: goList,
	}

	filePath, err := p.runTestCoverage()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(filePath)
	}()

	return p.parseCoverageProfile(filePath)
}

// parseCoverageData reads the coverage profile from the provided reader and
// returns it as a [Profile] struct.
func (p *parser) parseCoverageData(r io.Reader) (*Profile, error) {
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

	return p.parseLines(lines)
}

// parseCoverageProfile reads the coverage profile from the specified file path,
// parses it, and returns it as a [Profile] struct.
func (p *parser) parseCoverageProfile(filePath string) (*Profile, error) {
	var err error
	p.RawCoverData, err = os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return p.parseCoverageData(bytes.NewReader(p.RawCoverData))
}

// parseLines processes the lines of a Go coverage profile and returns them in a
// [Profile] struct.
func (p *parser) parseLines(lines []string) (*Profile, error) {
	if len(lines) < 2 {
		return nil, ErroInvalidCoverageFile
	}

	profile := NewProfile(p.RawGoTestOutput)

	if strings.HasPrefix(lines[0], "mode: ") {
		profile.Mode = Mode(strings.TrimPrefix(lines[0], "mode: "))
	}

	for _, line := range lines[1:] {
		var coverageLine Line

		// split at last ':' so paths that contain ':' (Windows drive letters) still work
		i := strings.LastIndex(line, ":")
		if i == -1 {
			return nil, fmt.Errorf("invalid coverage line: %q", line)
		}

		path := strings.TrimPrefix(line[:i], p.GoList.Module+"/")
		ranges := line[i+1:]

		_, err := fmt.Sscanf(ranges, "%d.%d,%d.%d %d %d",
			&coverageLine.StartLine, &coverageLine.StartColumn,
			&coverageLine.EndLine, &coverageLine.EndColumn,
			&coverageLine.Statements, &coverageLine.Count,
		)
		if err != nil {
			return nil, err
		}

		coverageLine.Path = path
		coverageLine.FilePath = filepath.Join(p.GoList.Dir, path)
		profile.Lines = append(profile.Lines, coverageLine)

		profile.expandCoveredLines(coverageLine)
	}

	return profile, nil
}

// runTestCoverage executes `go test -coverprofile` to generate a coverage
// profile file. It is the caller's responsibility to remove the file when it is
// no longer needed.
func (p *parser) runTestCoverage() (string, error) {
	tempDir, err := os.MkdirTemp("", "uncloak-*")
	if err != nil {
		return "", err
	}

	tempCoverFile := filepath.Join(tempDir, "coverage.out")

	cmd := exec.Command(
		"go",
		"test",
		"-coverprofile="+tempCoverFile,
		"./...",
	)

	p.RawGoTestOutput, err = cmd.CombinedOutput()
	if err != nil {
		return "", handleParseError(cmd, p.RawGoTestOutput, err)
	}

	return tempCoverFile, nil
}
