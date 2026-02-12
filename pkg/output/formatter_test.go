package output

import (
	"bytes"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	f := NewFormatter()
	if f == nil {
		t.Fatal("NewFormatter should return non-nil")
	}
}

func TestPrintVersionTable(t *testing.T) {
	f := NewFormatter()
	buf := bytes.NewBufferString("")
	f.SetOutput(buf)

	versions := []VersionItem{
		{Version: "v1.0.0", Size: "12.5MB", Date: "2026-02-12", Commit: "abc123", URL: "https://example.com"},
		{Version: "v1.0.1", Size: "13.2MB", Date: "2026-02-13", Commit: "def456", URL: "https://example.com/v1.0.1"},
	}

	f.PrintVersionTable(versions)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("v1.0.0")) {
		t.Error("output should contain version v1.0.0")
	}
}

func TestPrintJSON(t *testing.T) {
	f := NewFormatter()
	buf := bytes.NewBufferString("")
	f.SetOutput(buf)

	versions := []VersionItem{
		{Version: "v1.0.0", Size: "12.5MB", Date: "2026-02-12", Commit: "abc123", URL: "https://example.com"},
	}

	f.PrintJSON(versions)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("v1.0.0")) {
		t.Error("output should contain version v1.0.0")
	}
}
