package main

import "testing"

func TestDo(t *testing.T) {
	do := todo{
		command: "false",
	}

	err := do.Do()
	if err == nil {
		t.Fatalf("Do() failed to detect error")
	}

	do.command = "true"

	err = do.Do()
	if err != nil {
		t.Fatalf("Do() returned unexpected error: %s", err.Error())
	}
}
