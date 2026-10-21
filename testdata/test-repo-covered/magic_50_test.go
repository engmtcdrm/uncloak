package magic

import "testing"

func Test_50_add(t *testing.T) {
	expect := 5
	result := add(2, 3)

	if result != expect {
		t.Errorf("Expected %d, but got %d", expect, result)
	}
}
