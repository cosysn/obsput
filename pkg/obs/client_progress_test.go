package obs

import (
	"os"
	"testing"
)

func TestUploadWithProgressCallback(t *testing.T) {
	// Skip if endpoint is not reachable (requires real OBS)
	// This test requires a real OBS endpoint to test
	t.Skip("Skipping - requires real OBS endpoint")

	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test-upload-*.bin")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some test data
	testData := make([]byte, 1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	tmpFile.Write(testData)
	tmpFile.Close()

	client := NewClient("obs.test.com", "bucket", "ak", "sk")

	var progressCalled bool
	var received int64

	result, err := client.UploadFile(tmpFile.Name(), "v1.0.0", "", func(transferred int64) {
		progressCalled = true
		received = transferred
	})

	if err != nil {
		t.Fatalf("upload returned error: %v", err)
	}

	if !result.Success {
		t.Error("upload should succeed for test")
	}

	if !progressCalled {
		t.Error("progress callback should have been called")
	}

	if received <= 0 {
		t.Error("progress callback should report transferred bytes")
	}

	if received != 1024 {
		t.Errorf("expected 1024 bytes transferred, got %d", received)
	}
}
