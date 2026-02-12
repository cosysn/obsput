package cmd

import (
	"bytes"
	"testing"
)

func TestOBSCommand(t *testing.T) {
	cmd := NewOBSCommand()
	if cmd.Use != "obs" {
		t.Errorf("expected use 'obs', got '%s'", cmd.Use)
	}
}

func TestOBSCommandNew(t *testing.T) {
	cmd := NewOBSCommand()
	if len(cmd.Commands()) == 0 {
		t.Error("obs command should have subcommands")
	}
}

func TestOBSAddCommand(t *testing.T) {
	cmd := NewOBSAddCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute add --help failed: %v", err)
	}
}

func TestNewOBSInitCommand(t *testing.T) {
	cmd := NewOBSInitCommand()
	if cmd.Use != "init" {
		t.Errorf("expected use 'init', got '%s'", cmd.Use)
	}
}
