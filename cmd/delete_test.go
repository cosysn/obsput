package cmd

import (
	"bytes"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	cmd := NewDeleteCommand()
	if cmd.Use != "delete <version|--before>" {
		t.Errorf("expected use 'delete <version|--before>', got '%s'", cmd.Use)
	}
}

func TestDeleteCommandExecution(t *testing.T) {
	cmd := NewDeleteCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute delete --help failed: %v", err)
	}
}
