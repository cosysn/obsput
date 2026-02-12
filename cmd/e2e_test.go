package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestE2E_UploadToMinIO(t *testing.T) {
	// Skip if MINIO_ENDPOINT not set
	if os.Getenv("MINIO_ENDPOINT") == "" {
		t.Skip("MINIO_ENDPOINT not set, skipping E2E test")
	}

	// Create test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")
	content := []byte("test content for E2E upload")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("create test file failed: %v", err)
	}

	// Set environment for config
	os.Setenv("HOME", tmpDir)
	configPath := filepath.Join(tmpDir, ".obsput.yaml")
	configContent := `configs:
  - name: minio-test
    endpoint: "` + os.Getenv("MINIO_ENDPOINT") + `"
    bucket: "test-bucket"
    ak: "` + os.Getenv("MINIO_AK") + `"
    sk: "` + os.Getenv("MINIO_SK") + `"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	// Build binary
	cmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "obsput"), ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Run upload
	uploadCmd := exec.Command(filepath.Join(tmpDir, "obsput"), "upload", testFile)
	uploadCmd.Env = append(os.Environ(),
		"HOME="+tmpDir,
	)
	output, err := uploadCmd.CombinedOutput()

	if err != nil {
		t.Fatalf("upload failed: %v\n%s", err, string(output))
	}

	// Verify output contains success
	if !bytes.Contains(output, []byte("Uploaded:")) {
		t.Errorf("expected success message in output:\n%s", string(output))
	}

	// Verify output contains URL
	if !bytes.Contains(output, []byte("http://")) {
		t.Errorf("expected URL in output:\n%s", string(output))
	}
}

func TestE2E_ListFromMinIO(t *testing.T) {
	if os.Getenv("MINIO_ENDPOINT") == "" {
		t.Skip("MINIO_ENDPOINT not set, skipping E2E test")
	}

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Build binary
	cmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "obsput"), ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Run list
	listCmd := exec.Command(filepath.Join(tmpDir, "obsput"), "list")
	listCmd.Env = append(os.Environ(),
		"HOME="+tmpDir,
	)
	output, err := listCmd.CombinedOutput()

	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, string(output))
	}

	t.Logf("List output:\n%s", string(output))
}

func TestE2E_DeleteFromMinIO(t *testing.T) {
	if os.Getenv("MINIO_ENDPOINT") == "" {
		t.Skip("MINIO_ENDPOINT not set, skipping E2E test")
	}

	version := "v1.0.0-test123-20260101-120000"
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Build binary
	cmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "obsput"), ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Run delete
	deleteCmd := exec.Command(filepath.Join(tmpDir, "obsput"), "delete", version)
	deleteCmd.Env = append(os.Environ(),
		"HOME="+tmpDir,
	)
	output, err := deleteCmd.CombinedOutput()

	if err != nil {
		t.Fatalf("delete failed: %v\n%s", err, string(output))
	}

	if !bytes.Contains(output, []byte("Deleted:")) {
		t.Errorf("expected success message in output:\n%s", string(output))
	}
}
