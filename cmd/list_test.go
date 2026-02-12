package cmd

import (
	"bytes"
	"testing"
)

func TestListCommand(t *testing.T) {
	cmd := NewListCommand()
	if cmd.Use != "list" {
		t.Errorf("expected use 'list', got '%s'", cmd.Use)
	}
}

func TestListCommandExecution(t *testing.T) {
	cmd := NewListCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute list --help failed: %v", err)
	}
}
