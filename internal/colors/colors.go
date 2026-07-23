package colors

import pp "github.com/engmtcdrm/go-prettyprint"

// LightGreen returns the input text formatted with a light green color.
func LightGreen(text string) string {
	return pp.Fg24Bit(169, 209, 61, text)
}

// Green returns the input text formatted with a green color.
func Green(text string) string {
	return pp.Fg24Bit(97, 179, 79, text)
}

// MediumGreen returns the input text formatted with a medium green color.
func MediumGreen(text string) string {
	return pp.Fg24Bit(62, 139, 69, text)
}

// DarkGreen returns the input text formatted with a dark green color.
func DarkGreen(text string) string {
	return pp.Fg24Bit(51, 87, 59, text)
}
