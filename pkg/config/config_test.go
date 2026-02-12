package config

import (
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig should return non-nil")
	}
	if cfg.Configs == nil {
		t.Error("Configs should be initialized")
	}
}

func TestAddOBS(t *testing.T) {
	cfg := NewConfig()
	cfg.AddOBS("test-obs", "obs.test.com", "bucket", "ak", "sk")

	if len(cfg.Configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(cfg.Configs))
	}

	obs := cfg.Configs["test-obs"]
	if obs == nil {
		t.Fatal("config should exist")
	}
	if obs.Name != "test-obs" {
		t.Errorf("expected name 'test-obs', got '%s'", obs.Name)
	}
	if obs.Endpoint != "obs.test.com" {
		t.Errorf("expected endpoint 'obs.test.com', got '%s'", obs.Endpoint)
	}
}

func TestGetOBS(t *testing.T) {
	cfg := NewConfig()
	cfg.AddOBS("prod", "obs.prod.com", "bucket-prod", "ak1", "sk1")

	obs := cfg.GetOBS("prod")
	if obs == nil {
		t.Fatal("GetOBS should return non-nil")
	}
	if obs.Endpoint != "obs.prod.com" {
		t.Errorf("expected endpoint 'obs.prod.com', got '%s'", obs.Endpoint)
	}
}

func TestListOBS(t *testing.T) {
	cfg := NewConfig()
	cfg.AddOBS("obs1", "obs1.com", "bucket1", "ak", "sk")
	cfg.AddOBS("obs2", "obs2.com", "bucket2", "ak", "sk")

	list := cfg.ListOBS()
	if len(list) != 2 {
		t.Errorf("expected 2 configs, got %d", len(list))
	}
}

func TestRemoveOBS(t *testing.T) {
	cfg := NewConfig()
	cfg.AddOBS("to-remove", "obs.com", "bucket", "ak", "sk")
	cfg.RemoveOBS("to-remove")

	if len(cfg.Configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(cfg.Configs))
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")

	cfg := NewConfig()
	cfg.AddOBS("test", "obs.test.com", "bucket", "ak", "sk")

	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(loaded.Configs))
	}
}

func TestOBSExists(t *testing.T) {
	cfg := NewConfig()
	cfg.AddOBS("exists", "obs.com", "bucket", "ak", "sk")

	if !cfg.OBSExists("exists") {
		t.Error("Exists should return true for existing config")
	}
	if cfg.OBSExists("not-exists") {
		t.Error("Exists should return false for non-existing config")
	}
}
