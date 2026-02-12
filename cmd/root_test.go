package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestGetConfigPathFromBinary(t *testing.T) {
	// Test that getConfigPath returns path in binary directory
	path := getConfigPath()

	if path == "" {
		t.Error("getConfigPath should not return empty string")
	}

	if !filepath.IsAbs(path) {
		t.Error("getConfigPath should return absolute path")
	}

	// Should contain .obsput directory
	if !strings.Contains(path, ".obsput") {
		t.Errorf("path should contain .obsput, got: %s", path)
	}
}

func TestGetConfigPathNotInHome(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	path := getConfigPath()

	// Should NOT be in home directory
	if filepath.Dir(filepath.Dir(path)) == homeDir {
		t.Error("config path should not be in home directory")
	}
}
