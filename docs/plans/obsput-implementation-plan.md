# obsput CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a TDD-based CLI tool for uploading binaries to Huawei Cloud OBS with multi-OS support (Linux/Darwin/Windows)

**Architecture:**
- Cobra-based CLI with modular subcommands (upload, list, delete, download, obs)
- Viper for configuration management with YAML support
- Huawei Cloud OBS SDK for object storage operations
- Go-pretty for formatted table output
- TDD workflow: test → implement → refactor → commit

**Tech Stack:**
- Go 1.21+
- github.com/huaweicloud/huaweicloud-sdk-go-obs
- github.com/spf13/cobra
- github.com/spf13/viper
- github.com/jedib0t/go-pretty/v6

---

## Prerequisites

### Step 0.1: Initialize Go Module

**File:** N/A

**Step 1: Initialize go module**

Run: `go mod init obsput`
Expected: go.mod created

**Step 2: Add dependencies**

```bash
go get github.com/huaweicloud/huaweicloud-sdk-go-obs
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/jedib0t/go-pretty/v6
```

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: initialize go module and dependencies"
```

---

## Task 1: Project Structure & Root Command

**Files:**
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/obs.go`
- Create: `.gitignore`

**Step 1: Write failing test**

**File:** `cmd/root_test.go`

```go
package cmd

import (
	"bytes"
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestRootCommand -v`
Expected: FAIL - undefined: NewRootCommand

**Step 3: Write minimal implementation**

**File:** `main.go`

```go
package main

import "obsput/cmd"

func main() {
	cmd.Execute()
}
```

**File:** `cmd/root.go`

```go
package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"
var commit = "unknown"
var date = "unknown"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obsput",
		Short: "Upload binaries to Huawei Cloud OBS",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(NewOBSCommand())
	return cmd
}

func Execute() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestRootCommand -v`
Expected: PASS

**Step 5: Commit**

```bash
git add main.go cmd/root.go cmd/root_test.go .gitignore
git commit -m "feat: add root command structure"
```

---

## Task 2: Version Command & Build Info

**Files:**
- Create: `cmd/version.go`
- Modify: `cmd/root.go` (add version flag)

**Step 1: Write failing test**

**File:** `cmd/version_test.go`

```go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute version failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Version:") {
		t.Error("output should contain Version")
	}
}

func TestVersionFlag(t *testing.T) {
	root := NewRootCommand()
	buf := bytes.NewBufferString("")
	root.SetOut(buf)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute --version failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "obsput") {
		t.Error("output should contain obsput")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestVersion -v`
Expected: FAIL - undefined: NewVersionCommand

**Step 3: Write minimal implementation**

**File:** `cmd/version.go`

```go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Date: %s\n", date)
		},
	}
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewVersionCommand())
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestVersion -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/version.go cmd/version_test.go
git commit -m "feat: add version command"
```

---

## Task 3: OBS Config Package

**Files:**
- Create: `pkg/config/config.go`
- Create: `pkg/config/config_test.go`
- Create: `configs/obsput.yaml` (example)

**Step 1: Write failing test**

**File:** `pkg/config/config_test.go`

```go
package config

import (
	"os"
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/config/... -v`
Expected: FAIL - undefined: NewConfig, Load, OBS struct

**Step 3: Write minimal implementation**

**File:** `pkg/config/config.go`

```go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type OBS struct {
	Name     string `yaml:"name"`
	Endpoint string `yaml:"endpoint"`
	Bucket   string `yaml:"bucket"`
	AK       string `yaml:"ak"`
	SK       string `yaml:"sk"`
}

type Config struct {
	Configs map[string]*OBS `yaml:"configs"`
}

func NewConfig() *Config {
	return &Config{
		Configs: make(map[string]*OBS),
	}
}

func (c *Config) AddOBS(name, endpoint, bucket, ak, sk string) {
	c.Configs[name] = &OBS{
		Name:     name,
		Endpoint: endpoint,
		Bucket:   bucket,
		AK:       ak,
		SK:       sk,
	}
}

func (c *Config) GetOBS(name string) *OBS {
	return c.Configs[name]
}

func (c *Config) ListOBS() []*OBS {
	obsList := make([]*OBS, 0, len(c.Configs))
	for _, obs := range c.Configs {
		obsList = append(obsList, obs)
	}
	return obsList
}

func (c *Config) RemoveOBS(name string) {
	delete(c.Configs, name)
}

func (c *Config) OBSExists(name string) bool {
	_, ok := c.Configs[name]
	return ok
}

func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Configs == nil {
		cfg.Configs = make(map[string]*OBS)
	}
	return &cfg, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/config/config.go pkg/config/config_test.go configs/obsput.yaml
git commit -m "feat: add config package"
```

---

## Task 4: OBS Client Package

**Files:**
- Create: `pkg/obs/client.go`
- Create: `pkg/obs/client_test.go`

**Step 1: Write failing test**

**File:** `pkg/obs/client_test.go`

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/obs/... -v`
Expected: FAIL - undefined: NewClient

**Step 3: Write minimal implementation**

**File:** `pkg/obs/client.go`

```go
package obs

import (
	"fmt"
	"strings"
)

type Client struct {
	Endpoint string
	Bucket   string
	AK       string
	SK       string
}

func NewClient(endpoint, bucket, ak, sk string) *Client {
	return &Client{
		Endpoint: endpoint,
		Bucket:   bucket,
		AK:       ak,
		SK:       sk,
	}
}

func (c *Client) UploadURL(key string) string {
	return fmt.Sprintf("https://%s.%s/%s", c.Bucket, c.Endpoint, key)
}

func (c *Client) ParseVersionFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	return path
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/obs/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/obs/client.go pkg/obs/client_test.go
git commit -m "feat: add obs client package"
```

---

## Task 5: Version Generator Package

**Files:**
- Create: `pkg/version/generator.go`
- Create: `pkg/version/generator_test.go`

**Step 1: Write failing test**

**File:** `pkg/version/generator_test.go`

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/version/... -v`
Expected: FAIL - undefined: NewGenerator

**Step 3: Write minimal implementation**

**File:** `pkg/version/generator.go`

```go
package version

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate() string {
	commit := g.getShortCommit()
	now := time.Now()
	date := now.Format("20060102")
	timestamp := now.Format("150405")
	return fmt.Sprintf("v1.0.0-%s-%s-%s", commit, date, timestamp)
}

func (g *Generator) getShortCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/version/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/version/generator.go pkg/version/generator_test.go
git commit -m "feat: add version generator package"
```

---

## Task 6: Output Formatter Package

**Files:**
- Create: `pkg/output/formatter.go`
- Create: `pkg/output/formatter_test.go`

**Step 1: Write failing test**

**File:** `pkg/output/formatter_test.go`

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/output/... -v`
Expected: FAIL - undefined: NewFormatter, VersionItem

**Step 3: Write minimal implementation**

**File:** `pkg/output/formatter.go`

```go
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type VersionItem struct {
	Version string
	Size    string
	Date    string
	Commit  string
	URL     string
}

type Formatter struct {
	output io.Writer
}

func NewFormatter() *Formatter {
	return &Formatter{
		output: os.Stdout,
	}
}

func (f *Formatter) SetOutput(w io.Writer) {
	f.output = w
}

func (f *Formatter) PrintVersionTable(items []VersionItem) {
	t := table.NewWriter()
	t.SetOutputMirror(f.output)
	t.AppendHeader(table.Row{"VERSION", "SIZE", "DATE", "COMMIT", "DOWNLOAD_URL"})
	for _, item := range items {
		t.AppendRow(table.Row{item.Version, item.Size, item.Date, item.Commit, item.URL})
	}
	t.Render()
}

func (f *Formatter) PrintJSON(items []VersionItem) {
	enc := json.NewEncoder(f.output)
	enc.SetIndent("", "  ")
	enc.Encode(items)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/output/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/output/formatter.go pkg/output/formatter_test.go
git commit -m "feat: add output formatter package"
```

---

## Task 7: Upload Command

**Files:**
- Create: `cmd/upload.go`
- Create: `cmd/upload_test.go`

**Step 1: Write failing test**

**File:** `cmd/upload_test.go`

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestUploadCommand(t *testing.T) {
	cmd := NewUploadCommand()
	if cmd.Use != "upload" {
		t.Errorf("expected use 'upload', got '%s'", cmd.Use)
	}
}

func TestUploadCommandExecution(t *testing.T) {
	cmd := NewUploadCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute upload --help failed: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestUpload -v`
Expected: FAIL - undefined: NewUploadCommand

**Step 3: Write minimal implementation**

**File:** `cmd/upload.go`

```go
package cmd

import (
	"github.com/spf13/cobra"
)

func NewUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload binary to OBS",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("upload command")
		},
	}
	cmd.Flags().String("prefix", "", "Path prefix for upload")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewUploadCommand())
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestUpload -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/upload.go cmd/upload_test.go
git commit -m "feat: add upload command structure"
```

---

## Task 8: List Command

**Files:**
- Create: `cmd/list.go`
- Create: `cmd/list_test.go`

**Step 1: Write failing test**

**File:** `cmd/list_test.go`

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestListCommand(t *testing.T) {
	cmd := NewListCommand()
	if cmd.Use != "list" {
		t.Errorf("expected use 'list', got '%s'", cmd.Use)
	}
}

func TestListCommandExecution(t *testing.T) {
	cmd := NewListCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute list --help failed: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestList -v`
Expected: FAIL - undefined: NewListCommand

**Step 3: Write minimal implementation**

**File:** `cmd/list.go`

```go
package cmd

import (
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uploaded versions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("list command")
		},
	}
	cmd.Flags().StringP("output", "o", "table", "Output format (table/json)")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewListCommand())
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestList -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/list.go cmd/list_test.go
git commit -m "feat: add list command structure"
```

---

## Task 9: Delete Command

**Files:**
- Create: `cmd/delete.go`
- Create: `cmd/delete_test.go`

**Step 1: Write failing test**

**File:** `cmd/delete_test.go`

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	cmd := NewDeleteCommand()
	if cmd.Use != "delete" {
		t.Errorf("expected use 'delete', got '%s'", cmd.Use)
	}
}

func TestDeleteCommandExecution(t *testing.T) {
	cmd := NewDeleteCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute delete --help failed: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestDelete -v`
Expected: FAIL - undefined: NewDeleteCommand

**Step 3: Write minimal implementation**

**File:** `cmd/delete.go`

```go
package cmd

import (
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <version>",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("delete command")
		},
	}
	cmd.Flags().String("name", "", "OBS name to delete from")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewDeleteCommand())
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestDelete -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/delete.go cmd/delete_test.go
git commit -m "feat: add delete command structure"
```

---

## Task 10: Download Command

**Files:**
- Create: `cmd/download.go`
- Create: `cmd/download_test.go`

**Step 1: Write failing test**

**File:** `cmd/download_test.go`

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestDownloadCommand(t *testing.T) {
	cmd := NewDownloadCommand()
	if cmd.Use != "download" {
		t.Errorf("expected use 'download', got '%s'", cmd.Use)
	}
}

func TestDownloadCommandExecution(t *testing.T) {
	cmd := NewDownloadCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute download --help failed: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestDownload -v`
Expected: FAIL - undefined: NewDownloadCommand

**Step 3: Write minimal implementation**

**File:** `cmd/download.go`

```go
package cmd

import (
	"github.com/spf13/cobra"
)

func NewDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <version>",
		Short: "Show download commands for a version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("download command")
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output file path")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewDownloadCommand())
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestDownload -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/download.go cmd/download_test.go
git commit -m "feat: add download command structure"
```

---

## Task 11: OBS Config Command

**Files:**
- Create: `cmd/obs.go`
- Create: `cmd/obs_test.go`

**Step 1: Write failing test**

**File:** `cmd/obs_test.go`

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestOBSCommand(t *testing.T) {
	cmd := NewOBSCommand()
	if cmd.Use != "obs" {
		t.Errorf("expected use 'obs', got '%s'", cmd.Use)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestOBS -v`
Expected: FAIL - undefined: NewOBSCommand

**Step 3: Write minimal implementation**

**File:** `cmd/obs.go`

```go
package cmd

import (
	"github.com/spf13/cobra"
)

func NewOBSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obs",
		Short: "Manage OBS configurations",
	}
	cmd.AddCommand(NewOBSAddCommand())
	cmd.AddCommand(NewOBSListCommand())
	cmd.AddCommand(NewOBSGetCommand())
	cmd.AddCommand(NewOBSRemoveCommand())
	return cmd
}

func NewOBSAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add OBS configuration",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().String("name", "", "OBS name")
	cmd.Flags().String("endpoint", "", "OBS endpoint")
	cmd.Flags().String("bucket", "", "OBS bucket")
	cmd.Flags().String("ak", "", "Access Key")
	cmd.Flags().String("sk", "", "Secret Key")
	return cmd
}

func NewOBSListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OBS configurations",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return cmd
}

func NewOBSGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get OBS configuration",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return cmd
}

func NewOBSRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove OBS configuration",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return cmd
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestOBS -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/obs.go cmd/obs_test.go
git commit -m "feat: add obs config command structure"
```

---

## Task 12: Git Hook for Tests

**Files:**
- Create: `.githooks/pre-commit`

**Step 1: Write failing test**

**File:** N/A (verify hook doesn't exist)

Run: `test -f .githooks/pre-commit && echo "exists" || echo "not exists"`
Expected: not exists

**Step 2: Write implementation**

**File:** `.githooks/pre-commit`

```bash
#!/bin/bash
set -e

echo "Running tests..."
go test ./...

echo "All tests passed!"
```

Run: `chmod +x .githooks/pre-commit`

**Step 3: Verify**

Run: `test -f .githooks/pre-commit && echo "exists" || echo "not exists"`
Expected: exists

**Step 4: Commit**

```bash
git add .githooks/pre-commit
git commit -m "ci: add pre-commit test hook"
```

---

## Task 13: Makefile

**Files:**
- Create: `Makefile`

**Step 1: Write failing test**

**File:** N/A

Run: `test -f Makefile && echo "exists" || echo "not exists"`
Expected: not exists

**Step 2: Write implementation**

**File:** `Makefile`

```makefile
.PHONY: test build clean

test:
	go test ./...

build:
	go build -o obsput main.go

clean:
	rm -f obsput
```

**Step 3: Verify**

Run: `make test`
Expected: All tests pass

**Step 4: Commit**

```bash
git add Makefile
git commit -m "build: add Makefile"
```

---

## Task 14: README

**Files:**
- Create: `README.md`

**Step 1: Write implementation**

**File:** `README.md`

```markdown
# obsput

Upload binaries to Huawei Cloud OBS with CLI tool.

## Installation

```bash
go install
```

## Usage

```bash
# Upload
obsput upload ./bin/app --prefix releases

# List versions
obsput list

# Delete version
obsput delete v1.0.0-abc123-20260212-143000

# Download
obsput download v1.0.0-abc123-20260212-143000

# Configure OBS
obsput obs add --name prod --endpoint "obs.xxx.com" --bucket "bucket" --ak "ak" --sk "sk"
```
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README"
```

---

## Plan Complete

**Total Tasks:** 14

**Estimated Commits:** 14-16

**All tests should pass after each task**

Run: `go test ./...` at any point to verify.

---

## Plan Execution

**Plan complete and saved to `docs/plans/obsput-implementation-plan.md`. Two execution options:**

1. **Subagent-Driven (this session)** - Fresh subagent per task + code review
2. **Parallel Session (separate)** - New session uses superpowers:executing-plans

**Which approach?**
