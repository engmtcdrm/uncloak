package gocover

// parseMode validates and parses the coverage mode string.
func parseMode(mode string) (Mode, error) {
	switch Mode(mode) {
	case ModeSet, ModeCount, ModeAtomic:
		return Mode(mode), nil
	default:
		return "", ErrInvalidMode
	}
}
