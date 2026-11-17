package dummy

import (
	"runtime"
	"testing"
)

// Tests for [A] function.
func Test_A(t *testing.T) {
	t.Log("Running Test_A")

	A()
}

// Tests for [B] function.
func Test_B(t *testing.T) {
	t.Log("Running Test_B")

	B()
}

// Tests for [C] function.
func Test_C(t *testing.T) {
	t.Log("Running Test_C")

	C()
}

// Tests for [D] function.
func Test_D(t *testing.T) {
	t.Log("Running Test_D")

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("Skipping test on Windows and macOS")
	}

	D()
}

// // Tests for [E] function.
// func Test_E(_ *testing.T) {
// 	t.Log("Running Test_E")
//
// 	E()
// }
