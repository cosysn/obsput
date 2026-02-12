package obs

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")
	if client == nil {
		t.Fatal("NewClient should return non-nil")
	}
}

func TestUploadURL(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")

	url := client.UploadURL("path/to/file.bin")
	expected := "https://bucket.obs.test.com/path/to/file.bin"

	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestParseVersionFromPath(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")

	version := client.ParseVersionFromPath("v1.0.0-abc123-20260212-143000/")
	expected := "v1.0.0-abc123-20260212-143000"

	if version != expected {
		t.Errorf("expected %s, got %s", expected, version)
	}
}
