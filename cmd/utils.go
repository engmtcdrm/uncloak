package cmd

import (
	"regexp"
)

// getSemVer returns the semantic version of the input string if it
// matches the pattern `vX.Y.Z`. Otherwise, it returns the input string.
func getSemVer(input string) string {
	// Define the regular expression for semantic versioning
	re := regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)

	match := re.FindStringSubmatch(input)

	// If there's a match return the semantic version
	if len(match) > 1 {
		return match[1]
	}

	// If no match, return the original input
	return input
}
