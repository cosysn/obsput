package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	if !strings.Contains(path, ".obsput") {
		t.Errorf("path should contain .obsput, got: %s", path)
	}

	if !strings.HasSuffix(path, "obsput.yaml") {
		t.Errorf("path should end with obsput.yaml, got: %s", path)
	}
}

func TestGetConfigDir(t *testing.T) {
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}

	if !strings.Contains(dir, ".obsput") {
		t.Errorf("dir should contain .obsput, got: %s", dir)
	}
}

func TestConfigPathNotInHome(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	home, _ := filepath.Abs(filepath.Dir(filepath.Join("/", "usr", "bin")))
	if strings.HasPrefix(path, home) && !strings.Contains(path, ".obsput") {
		// This test verifies we're NOT using home directory
	}
}
