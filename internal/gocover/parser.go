package gocover

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

type parser struct {
	Options         *Options
	GoList          *GoList
	RawCoverData    []byte
	RawGoTestOutput []byte
	Command         string
}

// Run executes the 'go test' command to generate a coverage profile at the
// specified filePath. If filePath is empty, a temporary file will be created.
// Finally, it parses the coverage profile and returns it as a [Report].
func Run(ctx context.Context, filePath string, opts *Options) (*Report, error) {
	if opts == nil {
		opts = &DefaultOptions
	}

	goList, err := getGoList(ctx)
	if err != nil {
		return nil, err
	}

	p := &parser{
		Options: opts,
		GoList:  goList,
	}

	if filePath == "" {
		filePath, err = p.runTestCoverage(ctx)
		if err != nil {
			return NewReport(p.Command, p.RawGoTestOutput), err
		}
		defer os.Remove(filePath) //nolint:errcheck
	}

	return p.parseProfiles(filePath)
}

// parseProfiles reads the coverage profiles from the specified file path,
// parses it, and returns it as a [Report] struct.
func (p *parser) parseProfiles(fileName string) (*Report, error) {
	var err error
	p.RawCoverData, err = os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToReadFile, err)
	}

	if len(p.RawCoverData) == 0 {
		return nil, fmt.Errorf("%w: %w", ErrEmptyCoverageFile, io.EOF)
	}

	r := bytes.NewReader(p.RawCoverData)
	profiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCoverageFile, err)
	}

	return p.parseProfilesBlocks(profiles)
}

// parseProfilesBlocks parses the coverage profiles from the provided slice of
// [cover.Profile] and returns it as a [Report].
func (p *parser) parseProfilesBlocks(profiles []*cover.Profile) (*Report, error) {
	if profiles == nil {
		return nil, fmt.Errorf("%w: %w", ErrEmptyCoverageFile, io.EOF)
	}

	report := NewReport(p.Command, p.RawGoTestOutput)

	for _, profile := range profiles {
		path := strings.TrimPrefix(profile.FileName, p.GoList.Module+"/")

		cProfile := &Profile{
			Profile:          *profile,
			NoModuleFileName: path,
		}

		report.Profiles = append(report.Profiles, cProfile)

		for _, b := range profile.Blocks {
			report.expandCoveredLines(cProfile.NoModuleFileName, b)
		}
	}

	return report, nil
}

// runTestCoverage executes `go test -coverprofile` to generate a coverage
// profile file. It is the caller's responsibility to remove the file when it is
// no longer needed.
func (p *parser) runTestCoverage(ctx context.Context) (string, error) {
	tempDir, err := os.MkdirTemp("", "uncloak-*")
	if err != nil {
		return "", err
	}

	tempCoverFile := filepath.Join(tempDir, "coverage.out")

	cmd := exec.CommandContext(ctx, "go", "test")

	args := optionsToArgs(p.Options)
	cmd.Args = append(cmd.Args, args...)
	cmd.Args = append(cmd.Args,
		"-coverprofile="+tempCoverFile,
		"./...",
	)

	p.Command = strings.Join(cmd.Args, " ")

	p.RawGoTestOutput, err = cmd.CombinedOutput()
	if err != nil {
		return "", handleParseError(cmd, p.RawGoTestOutput, err)
	}

	return tempCoverFile, nil
}
