package magic

import "testing"

// These tests provide 100% test coverage

func Test_100_main(t *testing.T) {
	expected := 5
	AddC(2, 3)

	if c != expected {
		t.Errorf("Expected c to be %d, but got %d", expected, c)
	}
}

func Test_100_add(t *testing.T) {
	expect := 5
	result := add(2, 3)

	if result != expect {
		t.Errorf("Expected %d, but got %d", expect, result)
	}
}
