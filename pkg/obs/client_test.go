package obs

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")
	if client == nil {
		t.Fatal("NewClient should return non-nil")
	}
	if client.Endpoint != "obs.test.com" {
		t.Errorf("expected endpoint 'obs.test.com', got '%s'", client.Endpoint)
	}
}

func TestUploadURL(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")

	url := client.GetDownloadURL("path/to/file.bin")
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

func TestOBSClientEndpoint(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")
	if client.Endpoint != "obs.test.com" {
		t.Errorf("expected endpoint 'obs.test.com', got '%s'", client.Endpoint)
	}
}

func TestUploadKey(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")

	key := client.GetUploadKey("prefix", "v1.0.0-abc123-20260212-143000", "myapp")
	expected := "prefix/v1.0.0-abc123-20260212-143000/myapp"

	if key != expected {
		t.Errorf("expected %s, got %s", expected, key)
	}
}

func TestMD5Calculation(t *testing.T) {
	client := NewClient("obs.test.com", "bucket", "ak", "sk")

	// Create a temp file for testing
	content := []byte("test content")
	md5 := client.CalculateMD5(content)

	if md5 == "" {
		t.Error("MD5 should not be empty")
	}
	// MD5 is 16 bytes, base64 encoded is 24 characters
	if len(md5) != 24 {
		t.Errorf("MD5 should be 24 chars (base64), got %d", len(md5))
	}
	// Verify the expected base64 value for "test content"
	expected := "lHP90NiApDwht3eNNIchVw=="
	if md5 != expected {
		t.Errorf("expected %s, got %s", expected, md5)
	}
}

func TestDownloadURL(t *testing.T) {
	client := NewClient("obs.cn-east-1.myhuaweicloud.com", "my-bucket", "ak", "sk")

	url := client.GetDownloadURL("path/to/file.bin")
	expected := "https://my-bucket.obs.cn-east-1.myhuaweicloud.com/path/to/file.bin"

	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}
