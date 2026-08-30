package cmd

import (
	"errors"
	"fmt"

	pp "github.com/engmtcdrm/go-prettyprint"
)

var (
	errGitTargetRef = errors.New("flag -t/--target-ref is required. This should be the target ref to compare against, e.g. 'main'")
)

// coverageThresholdError indicates that the new code coverage is below the
// minimum required threshold.
type coverageThresholdError struct {
	coveragePercent   float64
	coverageThreshold float64
}

// newCoverageThresholdError creates a new [coverageThresholdError] with the
// given coverage percent and threshold.
func newCoverageThresholdError(coveragePercent, coverageThreshold float64) *coverageThresholdError {
	return &coverageThresholdError{
		coveragePercent:   coveragePercent,
		coverageThreshold: coverageThreshold,
	}
}

func (e *coverageThresholdError) Error() string {
	return fmt.Sprintf("new code coverage %s is below the minimum required %s",
		pp.Redf(floatFormat, e.coveragePercent),
		pp.Redf(floatFormat, e.coverageThreshold),
	)
}
