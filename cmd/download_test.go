package cmd

import (
	"bytes"
	"testing"
)

func TestDownloadCommand(t *testing.T) {
	cmd := NewDownloadCommand()
	if cmd.Use != "download <version>" {
		t.Errorf("expected use 'download <version>', got '%s'", cmd.Use)
	}
}

func TestDownloadCommandExecution(t *testing.T) {
	cmd := NewDownloadCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute download --help failed: %v", err)
	}
}
