package cmd

import (
	"regexp"
)

const semVerPattern = `^v?(\d+\.\d+\.\d+)$`

var semVerRegex = regexp.MustCompile(semVerPattern)

// getSemVer returns the semantic version of the input string if it
// matches the pattern `vX.Y.Z`. Otherwise, it returns the input string.
func getSemVer(input string) string {
	match := semVerRegex.FindStringSubmatch(input)

	if len(match) < 2 {
		return input
	}

	return match[1]
}
