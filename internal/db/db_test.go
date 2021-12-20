package db

import (
	"os"
	"path"
	"testing"
)

func TestReadConfig(t *testing.T) {
	tmp := t.TempDir()
	expected := "postgres://username:password@localhost:5432/database_name"
	tmpfile := path.Join(tmp, "dbconfig")
	if err := os.WriteFile(tmpfile, []byte(expected+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := ReadConfig(tmpfile)
	if err != nil {
		t.Fatal(err)
	}

	if config != expected {
		t.Fatalf("expected %q, got %q", expected, config)
	}
}

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
