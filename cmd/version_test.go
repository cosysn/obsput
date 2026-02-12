package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute version failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Version:") {
		t.Error("output should contain Version")
	}
}

func TestVersionFlag(t *testing.T) {
	root := NewRootCommand()
	buf := bytes.NewBufferString("")
	root.SetOut(buf)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute --version failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "obsput") {
		t.Error("output should contain obsput")
	}
}
