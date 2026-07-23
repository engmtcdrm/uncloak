package magic

var c int

func AddC(a, b int) {
	c = add(a, b)
}

func add(a, b int) int {
	return a + b
}
