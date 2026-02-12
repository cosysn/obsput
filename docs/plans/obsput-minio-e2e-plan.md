# MinIO E2E Test Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Use MinIO as Huawei Cloud OBS compatible endpoint for end-to-end testing

**Architecture:**
- Run MinIO container for OBS simulation
- Create test configuration pointing to MinIO
- Test upload, list, delete operations
- Verify OBS SDK compatibility with MinIO (both use S3 protocol)

**Tech Stack:**
- MinIO Server (Docker)
- github.com/minio/minio-go or huaweicloud SDK with custom endpoint

---

## Task 1: Create Docker Compose for MinIO

**Files:**
- Create: `docker-compose.yaml`
- Create: `scripts/start-minio.sh`

**Step 1: Write implementation**

**File:** `docker-compose.yaml`

```yaml
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    container_name: obsput-minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: admin
      MINIO_ROOT_PASSWORD: password
    volumes:
      - minio-data:/data
    command: server --console-address ":9001" /data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

volumes:
  minio-data:
```

**File:** `scripts/start-minio.sh`

```bash
#!/bin/bash
set -e

echo "Starting MinIO..."
docker-compose up -d

echo "Waiting for MinIO to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:9000/minio/health/live > /dev/null 2>&1; then
        echo "MinIO is ready!"
        echo "Console: http://localhost:9001 (admin/password)"
        echo "API: http://localhost:9000"
        exit 0
    fi
    sleep 1
done

echo "MinIO failed to start"
exit 1
```

**Step 2: Run and verify**

Run: `chmod +x scripts/start-minio.sh && ./scripts/start-minio.sh`
Expected: MinIO starts successfully

**Step 3: Commit**

```bash
git add docker-compose.yaml scripts/start-minio.sh
git commit -m "feat: add MinIO docker compose"
```

---

## Task 2: Create Test Configuration

**Files:**
- Create: `testdata/minio-config.yaml`

**Step 1: Write test configuration**

**File:** `testdata/minio-config.yaml`

```yaml
configs:
  - name: minio-local
    endpoint: "localhost:9000"
    bucket: "test-bucket"
    ak: "admin"
    sk: "password"
```

**Step 2: Create bucket on startup**

**File:** `scripts/setup-minio.sh`

```bash
#!/bin/bash

# Wait for MinIO to be ready
./scripts/start-minio.sh

# Create test bucket
mc alias set myminio http://localhost:9000 admin password || true
mc mb myminio/test-bucket --ignore-existing

echo "Test bucket created!"
```

**Step 3: Commit**

```bash
git add testdata/minio-config.yaml scripts/setup-minio.sh
git commit -m "feat: add test configuration"
```

---

## Task 3: Create E2E Test Suite

**Files:**
- Create: `cmd/e2e_test.go`

**Step 1: Write E2E tests**

**File:** `cmd/e2e_test.go`

```go
package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if !bytes.Contains(output, []byte("✓ Uploaded:")) {
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

	if !bytes.Contains(output, []byte("✓ Deleted:")) {
		t.Errorf("expected success message in output:\n%s", string(output))
	}
}
```

**Step 2: Run tests (should skip without env vars)**

Run: `go test ./cmd/... -run TestE2E -v`
Expected: SKIP (MINIO_ENDPOINT not set)

**Step 3: Commit**

```bash
git add cmd/e2e_test.go
git commit -m "test: add E2E test suite"
```

---

## Task 4: Add Makefile E2E Target

**Files:**
- Modify: `Makefile`

**Step 1: Add E2E target**

**File:** `Makefile`

```makefile
.PHONY: test build clean release all e2e e2e-test

# ... existing content ...

e2e:
	MINIO_ENDPOINT=localhost:9000 \
	MINIO_AK=admin \
	MINIO_SK=password \
	go test ./cmd/... -run TestE2E -v

e2e-setup: docker-compose.yaml scripts/start-minio.sh
	./scripts/start-minio.sh
	mc alias set myminio http://localhost:9000 admin password
	mc mb myminio/test-bucket --ignore-existing

e2e-clean:
	docker-compose down -v
```

**Step 2: Test E2E target (should skip)**

Run: `make e2e`
Expected: Tests skipped (no MinIO running)

**Step 3: Commit**

```bash
git add Makefile
git commit -m "test: add E2E make targets"
```

---

## Task 5: Test with Real MinIO

**Step 1: Start MinIO**

Run: `make e2e-setup`
Expected: MinIO starts, bucket created

**Step 2: Run E2E tests**

Run: `make e2e`
Expected: All E2E tests pass

**Step 3: Commit**

```bash
git add -a
git commit -m "test: run E2E tests with MinIO"
```

---

## Task 6: Verify OBS SDK Compatibility

**Files:**
- Modify: `pkg/obs/client.go`

**Step 1: Ensure SDK works with MinIO**

MinIO is S3-compatible. The Huawei Cloud SDK may need configuration changes.

**Step 2: Add compatibility note**

```go
// Note: Huawei Cloud OBS SDK is compatible with S3-compatible services
// including MinIO. Set the endpoint to your MinIO server URL.
```

**Step 3: Commit**

```bash
git add pkg/obs/client.go
git commit -m "docs: note S3 compatibility"
```

---

## Plan Complete

**Total Tasks:** 6

**Goal achieved:**
- MinIO container for OBS simulation
- E2E test suite for upload, list, delete
- Make targets for test automation
- CI/CD ready test setup

---

**Plan complete and saved to `docs/plans/obsput-minio-e2e-plan.md`**

**Execution option:**
- Use superpowers:subagent-driven-development to execute task-by-task with code review
