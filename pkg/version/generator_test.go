package version

import (
	"strings"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Fatal("NewGenerator should return non-nil")
	}
}

func TestGenerateVersion(t *testing.T) {
	g := NewGenerator()
	version := g.Generate()

	if version == "" {
		t.Fatal("Generate should return non-empty string")
	}

	// Check format: v<version>-<commit>-<date>-<time>
	parts := strings.Split(version, "-")
	if len(parts) < 4 {
		t.Errorf("version should have at least 4 parts, got %d: %s", len(parts), version)
	}
}

func TestVersionFormat(t *testing.T) {
	g := NewGenerator()
	version := g.Generate()

	// Should start with 'v'
	if !strings.HasPrefix(version, "v") {
		t.Errorf("version should start with 'v', got '%s'", version)
	}

	// Date should be YYYYMMDD format
	datePart := strings.Split(version, "-")[2]
	if len(datePart) != 8 {
		t.Errorf("date part should be 8 chars, got '%s'", datePart)
	}

	// Time should be HHMMSS format
	timePart := strings.Split(version, "-")[3]
	if len(timePart) != 6 {
		t.Errorf("time part should be 6 chars, got '%s'", timePart)
	}
}

func TestVersionConsistency(t *testing.T) {
	g := NewGenerator()
	// Same second should produce same version
	v1 := g.Generate()
	time.Sleep(10 * time.Millisecond)
	v2 := g.Generate()

	if v1 == v2 {
		t.Error("different calls should produce different versions")
	}
}
