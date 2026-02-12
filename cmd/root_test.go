package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	root := NewRootCommand()
	if root == nil {
		t.Fatal("NewRootCommand should return non-nil command")
	}
	if root.Name() != "obsput" {
		t.Errorf("expected name 'obsput', got '%s'", root.Name())
	}
}

func TestRootCommandVersion(t *testing.T) {
	root := NewRootCommand()
	buf := bytes.NewBufferString("")
	root.SetOut(buf)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute version failed: %v", err)
	}
}
