package cmd

import (
	"bytes"
	"testing"
)

func TestUploadCommand(t *testing.T) {
	cmd := NewUploadCommand()
	if cmd.Use != "upload <file>" {
		t.Errorf("expected use 'upload <file>', got '%s'", cmd.Use)
	}
}

func TestUploadCommandExecution(t *testing.T) {
	cmd := NewUploadCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute upload --help failed: %v", err)
	}
}
