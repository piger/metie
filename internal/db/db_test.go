package db

import (
	"testing"
)

func TestMakeColumnString(t *testing.T) {
	names := []string{
		"one",
		"two",
		"three",
	}

	expected := "one,two,three"
	result := makeColumnString(names)
	if expected != result {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestMakeValuesString(t *testing.T) {
	names := []string{
		"one",
		"two",
		"three",
	}
	expected := "$1,$2,$3"
	result := makeValuesString(names)
	if expected != result {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}
