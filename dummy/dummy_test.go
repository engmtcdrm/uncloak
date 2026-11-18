package dummy

import (
	"runtime"
	"testing"
)

// Tests for [A] function.
func Test_A(t *testing.T) {
	t.Log("Testing function A")
	A()
}

// Tests for [B] function.
func Test_B(t *testing.T) {
	t.Log("Testing function B")
	B()
}

// Tests for [C] function.
func Test_C(t *testing.T) {
	t.Log("Testing function C")
	C()
}

// Tests for [D] function.
func Test_D(t *testing.T) {
	t.Log("Testing function D")
	D()
}

// Tests for [E] function.
func Test_E(t *testing.T) {
	t.Log("Testing function E")
	E()
}

// Tests for [F] function.
func Test_F(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	t.Log("Testing function F")
	F()
}
