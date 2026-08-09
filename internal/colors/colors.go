package colors

import pp "github.com/engmtcdrm/go-prettyprint"

// LightGreen returns the input text formatted with a light green color.
func LightGreen(a ...any) string {
	return pp.Fg24Bit(169, 209, 61, a...)
}

// Green returns the input text formatted with a green color.
func Green(a ...any) string {
	return pp.Fg24Bit(97, 179, 79, a...)
}

func BoldGreen(a ...any) string {
	return pp.Bold(Green(a...))
}

// MediumGreen returns the input text formatted with a medium green color.
func MediumGreen(a ...any) string {
	return pp.Fg24Bit(62, 139, 69, a...)
}

// DarkGreen returns the input text formatted with a dark green color.
func DarkGreen(a ...any) string {
	return pp.Fg24Bit(51, 87, 59, a...)
}
