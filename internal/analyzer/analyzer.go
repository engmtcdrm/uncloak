package analyzer

import (
	"errors"
	"fmt"
	"sync"

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

// processFiles reads and parses the Go coverage profile and the git diff file
// concurrently. It returns the parsed coverage profile and git diff, or an
// error if any occurs during parsing.
func processFiles(cfg *config.Config) (*gocover.Profile, *gitdiff.Results, error) {
	var wg sync.WaitGroup
	var errs error

	tm := task.NewManager()
	tm.Start()

	covCh := make(chan struct {
		profile *gocover.Profile
		err     error
	}, 1)
	diffCh := make(chan struct {
		diff *gitdiff.Results
		err  error
	}, 1)

	wg.Go(func() {
		gotask := task.NewTask("go", "Running Go coverage analysis")
		tm.AddTask(gotask)

		var err error

		gotask.Start()
		defer func() {
			gotask.SetMessage("Finished Go coverage analysis")

			switch {
			case err != nil:
				gotask.Error()
			default:
				gotask.Finish()
			}
		}()

		p, err := gocover.Run()
		covCh <- struct {
			profile *gocover.Profile
			err     error
		}{p, err}
	})

	wg.Go(func() {
		gittask := task.NewTask("git", "Running Git diff analysis")
		tm.AddTask(gittask)

		var err error

		gittask.Start()
		defer func() {
			gittask.SetMessage("Finished Git diff analysis")

			switch {
			case err != nil:
				gittask.Error()
			default:
				gittask.Finish()
			}
		}()

		d, err := gitdiff.Run(&cfg.GitDiffOptions)
		diffCh <- struct {
			diff *gitdiff.Results
			err  error
		}{d, err}
	})

	wg.Wait()
	tm.Finish()

	fmt.Println()

	covRes := <-covCh
	diffRes := <-diffCh

	printCommandsIfDebug(cfg, covRes.profile, diffRes.diff)

	errs = errors.Join(covRes.err, diffRes.err)
	if errs != nil {
		return nil, nil, errs
	}

	return covRes.profile, diffRes.diff, nil
}

// printCommandsIfDebug prints the commands used for Go coverage and Git diff
// analysis if debug mode is enabled.
func printCommandsIfDebug(cfg *config.Config, coverageProfile *gocover.Profile, diffResults *gitdiff.Results) {
	if !cfg.Debug {
		return
	}

	if coverageProfile != nil && coverageProfile.Command != "" {
		fmt.Printf("Go coverage analysis command ran: %s\n", pp.Cyan(coverageProfile.Command))
	}

	if diffResults != nil && diffResults.Command != "" {
		fmt.Printf("Git diff analysis command ran: %s\n\n", pp.Cyan(diffResults.Command))
	}
}
