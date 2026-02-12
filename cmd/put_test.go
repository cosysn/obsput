package cmd

import (
	"bytes"
	"testing"
)

func TestPutCommand(t *testing.T) {
	cmd := NewPutCommand()
	if cmd.Use != "put <file>" {
		t.Errorf("expected use 'put <file>', got '%s'", cmd.Use)
	}
}

func TestPutCommandExecution(t *testing.T) {
	cmd := NewPutCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute put --help failed: %v", err)
	}
}
