package analyzer

import (
	"errors"
	"sync"

	"github.com/engmtcdrm/uncloak/internal/config"
	"github.com/engmtcdrm/uncloak/internal/gitdiff"
	"github.com/engmtcdrm/uncloak/internal/gocover"
)

var ErrBelowThreshold = errors.New("new code coverage below threshold")

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

	// If there are no new lines in any Go files, exit early
	if len(report.GitDiffResults.NewLines) == 0 {
		return report, nil
	}

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
	var filteredFiles []string

	if len(cfg.Exclusions) > 0 {
		for _, file := range report.GitDiffResults.Files() {
			if !cfg.IsExclusionFile(file) {
				filteredFiles = append(filteredFiles, file)
			}
		}
	} else {
		filteredFiles = report.GitDiffResults.Files()
	}

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

// processFiles reads and parses the Go coverage profile and the git diff file
// concurrently. It returns the parsed coverage profile and git diff, or an
// error if any occurs during parsing.
func processFiles(cfg *config.Config) (*gocover.Profile, *gitdiff.Results, error) {
	var wg sync.WaitGroup
	var errs error

	covCh := make(chan struct {
		profile *gocover.Profile
		err     error
	}, 1)
	diffCh := make(chan struct {
		diff *gitdiff.Results
		err  error
	}, 1)

	wg.Go(func() {
		p, err := gocover.Run(cfg.Debug)
		covCh <- struct {
			profile *gocover.Profile
			err     error
		}{p, err}
	})

	wg.Go(func() {
		d, err := gitdiff.Run(cfg.Debug, &cfg.GitDiffOptions)
		diffCh <- struct {
			diff *gitdiff.Results
			err  error
		}{d, err}
	})

	wg.Wait()

	covRes := <-covCh
	diffRes := <-diffCh

	errs = errors.Join(covRes.err, diffRes.err)

	if errs != nil {
		return nil, nil, errs
	}

	return covRes.profile, diffRes.diff, nil
}
