package analyzer

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	pp "github.com/engmtcdrm/go-prettyprint"

	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/gocover"
	"github.com/engmtcdrm/uncloak/internal/task"
)

var (
	// ErrBelowThreshold is returned when the new code coverage is below the
	// configured threshold.
	ErrBelowThreshold = errors.New("new code coverage below threshold")
)

// NewCodeCoverage analyzes the Go coverage profile and the git diff file,
// returning a report of the coverage of new lines in the code. It returns
// a [Report] of the analyzed coverage and an error if any occurs during the
// analysis. If the coverage of new lines is below the threshold specified in
// the configuration, it returns [ErrBelowThreshold] as the error.
func NewCodeCoverage(cfg *config.Config) (*Report, error) {
	profile, diff, err := processFiles(cfg)
	if err != nil {
		return &Report{}, err
	}

	report := NewReport(cfg.CoverageThreshold, profile, diff)
	report = analyzeCoverage(report, cfg)

	// If there are no new lines, changes must have been outside of test
	// coverage lines.
	if report.TotalNewLines() == 0 {
		return report, nil
	}

	report.GroupUncoveredLines()
	coveragePercent := report.CoveragePercent()

	if coveragePercent < report.CoverageThreshold {
		return report, ErrBelowThreshold
	}

	return report, nil
}

// analyzeCoverage analyzes the coverage of new lines in the given report based
// on the configuration. It returns the updated report with the analyzed
// coverage.
func analyzeCoverage(report *Report, cfg *config.Config) *Report {
	filteredFiles := filterFiles(cfg, report.GitDiffResults.Files())

	for _, file := range filteredFiles {
		newLines := report.GitDiffResults.NewLines[file]

		if report.CoverageProfile.CoveredLines[file] == nil {
			continue
		}

		reportFile := NewFileReport(file)

		for line := range newLines {
			if !report.CoverageProfile.IsInTestCoverage(file, line) {
				continue
			}

			if report.CoverageProfile.CoveredLines[file][line] {
				reportFile.CoveredNewLines = append(reportFile.CoveredNewLines, line)
				continue
			}

			reportFile.UncoveredNewLines = append(reportFile.UncoveredNewLines, line)
		}

		report.Files = append(report.Files, reportFile)
	}

	return report
}

// filterFiles filters the given list of files based on the exclusions specified
// in the configuration. It returns a new list of files that are not excluded.
func filterFiles(cfg *config.Config, files []string) []string {
	if len(cfg.Exclusions) == 0 {
		return files
	}

	var filteredFiles []string

	for _, file := range files {
		if !cfg.IsExclusionFile(file) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles
}

// printCommands prints the commands used for Git diff and Go coverage analysis.
func printCommands(coverageProfile *gocover.Profile, diffResults *gitdiff.Results) {
	var buf bytes.Buffer

	if diffResults != nil && diffResults.Command != "" {
		fmt.Fprintf(&buf, "Git diff analysis command ran: %s\n", pp.Cyan(diffResults.Command))
	}

	if coverageProfile != nil && coverageProfile.Command != "" {
		fmt.Fprintf(&buf, "Go test coverage analysis command ran: %s\n", pp.Cyan(coverageProfile.Command))
	}

	if buf.Len() > 0 {
		fmt.Fprintf(&buf, "\n")
	}

	fmt.Print(buf.String())
}

// processFiles reads and parses the Go coverage profile and the git diff file
// concurrently. It returns the parsed coverage profile and git diff, or an
// error if any occurs during parsing.
func processFiles(cfg *config.Config) (*gocover.Profile, *gitdiff.Results, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tm := task.NewManager()
	tm.Start()

	diff, diffErr := runTaskGitDiff(ctx, tm, &cfg.GitDiffOptions)
	if diffErr != nil {
		tm.Finish()

		fmt.Println()

		if cfg.Debug {
			printCommands(nil, diff)
		}

		if errors.Is(diffErr, gitdiff.ErrNoOutput) {
			return nil, nil, nil
		}

		return nil, nil, diffErr
	}

	profile, profileErr := runTaskGoCoverage(ctx, tm, cfg.CoverageFile, &cfg.GoTestOptions)
	if profileErr != nil {
		tm.Finish()

		fmt.Println()

		if cfg.Debug {
			printCommands(profile, diff)
		}

		return nil, nil, profileErr
	}

	tm.Finish()

	fmt.Println()

	if cfg.Debug {
		printCommands(profile, diff)
	}

	return profile, diff, nil
}

// runTaskGitDiff executes the Git diff analysis task and returns the resulting
// diff results and any error encountered.
func runTaskGitDiff(ctx context.Context, tm *task.Manager, opts *gitdiff.Options) (*gitdiff.Results, error) {
	gittask := task.NewTask("git", "Running Git diff analysis")
	tm.AddTask(gittask)

	gittask.Start()

	d, err := gitdiff.Run(ctx, opts)

	gittask.SetMessage("Finished Git diff analysis")

	switch {
	case err != nil:
		switch {
		case errors.Is(err, gitdiff.ErrNoOutput):
			gittask.SetMessage("Finished Git diff analysis - No changes detected, no need to continue analysis")
			gittask.Finish()
		case errors.Is(ctx.Err(), context.Canceled):
			gittask.SetMessage("Git diff analysis stopped - Please see Go test coverage analysis as to why it was stopped")
			gittask.Warning()
			err = newTaskCanceledError(err)
		default:
			gittask.Error()
		}
	default:
		gittask.Finish()
	}

	return d, err
}

// runTaskGoCoverage executes the Go test coverage analysis task and returns the
// resulting coverage profile and any error encountered.
func runTaskGoCoverage(ctx context.Context, tm *task.Manager, coverageFilePath string, opts *gocover.Options) (*gocover.Profile, error) {
	gotask := task.NewTask("go", "Running Go test coverage analysis")
	tm.AddTask(gotask)

	gotask.Start()
	p, err := gocover.Run(ctx, coverageFilePath, opts)

	gotask.SetMessage("Finished Go test coverage analysis")

	switch {
	case err != nil:
		if errors.Is(ctx.Err(), context.Canceled) {
			gotask.SetMessage("Go test coverage analysis stopped - Please see Git diff analysis as to why it was stopped")
			gotask.Warning()
			err = newTaskCanceledError(err)
		} else {
			gotask.Error()
		}
	default:
		gotask.Finish()
	}

	return p, err
}
