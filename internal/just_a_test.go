package internal

import (
    "testing"
)

func TestShouldPass(t *testing.T) {
	if 1 != 1 {
		t.Fatalf("Something is wrong in the universe...")
	}
}

func TestShouldFail(t *testing.T) {
	//t.Fatalf("TestShouldFail failed.")
}