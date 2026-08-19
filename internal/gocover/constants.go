package gocover

// Mode represents the coverage mode used by the Go test coverage tool.
type Mode string

const (
	// ModeSet represents the set coverage mode.
	ModeSet Mode = "set"

	// ModeCount represents the count coverage mode.
	ModeCount Mode = "count"

	// ModeAtomic represents the atomic coverage mode.
	ModeAtomic Mode = "atomic"
)
