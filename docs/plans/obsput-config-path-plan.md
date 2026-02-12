# Config Path Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Change config path from ~/.obsput/obsput.yaml to {binary_dir}/.obsput/obsput.yaml with auto-generate template on first run

**Architecture:**
- Get binary path via os.Executable()
- Store config at {binary_dir}/.obsput/obsput.yaml
- Add embedded default config template
- Check for empty config and show user-friendly message

**Tech Stack:**
- os.Executable() for binary path
- embed.FS for config template (Go 1.16+)

---

## Task 1: Create Config Template File

**Files:**
- Create: `pkg/config/template.yaml`
- Create: `pkg/config/template.go`

**Step 1: Write the config template**

**File:** `pkg/config/template.yaml`

```yaml
configs: []
```

**Step 2: Write Go file to embed template**

**File:** `pkg/config/template.go`

```go
package config

import "embed"

//go:embed template.yaml
var FS embed.FS

const DefaultConfig = `configs: []
`
```

**Step 3: Test it works**

Run: `go build ./pkg/config/...`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add pkg/config/template.yaml pkg/config/template.go
git commit -m "feat: add config template file"
```

---

## Task 2: Update config.go with GetConfigPath

**Files:**
- Modify: `pkg/config/config.go`
- Create: `pkg/config/config_path_test.go`

**Step 1: Write failing test**

**File:** `pkg/config/config_path_test.go`

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/config/... -run TestGet -v`
Expected: FAIL - undefined: GetConfigPath

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

func GetConfigPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(execPath)
	return filepath.Join(dir, ".obsput", "obsput.yaml"), nil
}

func GetConfigDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(execPath), ".obsput"), nil
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

func (c *Config) Ensure(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.Save(path)
	}
	return nil
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

func LoadOrInit() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	cfg, err := Load(path)
	if err != nil {
		// Config doesn't exist, create new one
		cfg = NewConfig()
		// Auto-generate template
		if err := cfg.Ensure(path); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/config/... -run TestGet -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/config/config.go pkg/config/config_path_test.go
git commit -m "feat: add GetConfigPath function"
```

---

## Task 3: Update cmd/root.go

**Files:**
- Modify: `cmd/root.go`
- Create: `cmd/root_test.go`

**Step 1: Write failing test**

**File:** `cmd/root_test.go`

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPathFromBinary(t *testing.T) {
	// Test that getConfigPath returns path in binary directory
	path := getConfigPath()

	if path == "" {
		t.Error("getConfigPath should not return empty string")
	}

	if !filepath.IsAbs(path) {
		t.Error("getConfigPath should return absolute path")
	}

	// Should contain .obsput directory
	if !filepath.Contains(path, ".obsput") {
		t.Errorf("path should contain .obsput, got: %s", path)
	}
}

func TestGetConfigPathNotInHome(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	path := getConfigPath()

	// Should NOT be in home directory
	if filepath.Dir(filepath.Dir(path)) == homeDir {
		t.Error("config path should not be in home directory")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestGetConfig -v`
Expected: FAIL - undefined: getConfigPath

**Step 3: Write implementation**

**File:** `cmd/root.go`

```go
package cmd

import (
	"os"
	"path/filepath"

	"obsput/pkg/config"

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
	cmd.AddCommand(NewUploadCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewDownloadCommand())
	cmd.AddCommand(NewVersionCommand())
	return cmd
}

func Execute() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func getConfigPath() string {
	path, _ := config.GetConfigPath()
	return path
}

func getConfigDir() string {
	dir, _ := config.GetConfigDir()
	return dir
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestGetConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/root.go cmd/root_test.go
git commit -m "feat: update root command with new config path"
```

---

## Task 4: Update cmd/obs.go

**Files:**
- Modify: `cmd/obs.go`
- Modify: `cmd/obs_test.go`

**Step 1: Write failing test**

**File:** `cmd/obs_test.go`

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestOBSCommandNew(t *testing.T) {
	cmd := NewOBSCommand()
	if len(cmd.Commands()) == 0 {
		t.Error("obs command should have subcommands")
	}
}

func TestOBSAddCommand(t *testing.T) {
	cmd := NewOBSAddCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute add --help failed: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestOBS -v`
Expected: FAIL - getConfigPath already defined elsewhere

**Step 3: Write implementation**

**File:** `cmd/obs.go`

```go
package cmd

import (
	"fmt"
	"os"

	"obsput/pkg/config"

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
	cmd.AddCommand(NewOBSInitCommand())
	return cmd
}

func NewOBSAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add OBS configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			bucket, _ := cmd.Flags().GetString("bucket")
			ak, _ := cmd.Flags().GetString("ak")
			sk, _ := cmd.Flags().GetString("sk")

			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			cfg.AddOBS(name, endpoint, bucket, ak, sk)

			if err := cfg.Save(getConfigPath()); err != nil {
				return fmt.Errorf("save config failed: %v", err)
			}

			cmd.Printf("✓ Added OBS config: %s\n", name)
			return nil
		},
	}
	cmd.Flags().String("name", "", "OBS name")
	cmd.Flags().String("endpoint", "", "OBS endpoint")
	cmd.Flags().String("bucket", "", "OBS bucket")
	cmd.Flags().String("ak", "", "Access Key")
	cmd.Flags().String("sk", "", "Secret Key")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("bucket")
	cmd.MarkFlagRequired("ak")
	cmd.MarkFlagRequired("sk")
	return cmd
}

func NewOBSListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OBS configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			cmd.Println("NAME\tENDPOINT\tBUCKET\tSTATUS")
			for _, obs := range cfg.ListOBS() {
				cmd.Printf("%s\t%s\t%s\tactive\n", obs.Name, obs.Endpoint, obs.Bucket)
			}
			return nil
		},
	}
	return cmd
}

func NewOBSGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get OBS configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			obs := cfg.GetOBS(name)
			if obs == nil {
				return fmt.Errorf("OBS config not found: %s", name)
			}

			cmd.Printf("Name: %s\n", obs.Name)
			cmd.Printf("Endpoint: %s\n", obs.Endpoint)
			cmd.Printf("Bucket: %s\n", obs.Bucket)
			cmd.Printf("AK: %s\n", maskAK(obs.AK))
			cmd.Printf("SK: %s\n", maskSK(obs.SK))
			return nil
		},
	}
	return cmd
}

func NewOBSRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove OBS configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			if !cfg.OBSExists(name) {
				return fmt.Errorf("OBS config not found: %s", name)
			}

			cfg.RemoveOBS(name)

			if err := cfg.Save(getConfigPath()); err != nil {
				return fmt.Errorf("save config failed: %v", err)
			}

			cmd.Printf("✓ Removed OBS config: %s\n", name)
			return nil
		},
	}
	return cmd
}

func NewOBSInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize OBS configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getConfigPath()

			cfg := config.NewConfig()
			if err := cfg.Ensure(path); err != nil {
				return fmt.Errorf("init config failed: %v", err)
			}

			cmd.Printf("✓ Initialized config: %s\n", path)
			cmd.Println("\nAdd OBS configuration:")
			cmd.Println("  obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			return nil
		},
	}
	return cmd
}

func maskAK(ak string) string {
	if len(ak) <= 4 {
		return "****"
	}
	return ak[:len(ak)-4] + "****"
}

func maskSK(sk string) string {
	if len(sk) <= 4 {
		return "****"
	}
	return sk[:len(sk)-4] + "****"
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestOBS -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/obs.go cmd/obs_test.go
git commit -m "feat: update obs command with new config path"
```

---

## Task 5: Add getConfig helper to other commands

**Files:**
- Modify: `cmd/upload.go`
- Modify: `cmd/list.go`
- Modify: `cmd/delete.go`
- Modify: `cmd/download.go`

**Step 1: Write failing test**

**File:** `cmd/upload_test.go`

```go
package cmd

import (
	"bytes"
	"testing"
)

func TestUploadCommandHasConfigCheck(t *testing.T) {
	cmd := NewUploadCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help failed: %v", err)
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./cmd/... -run TestUpload -v`
Expected: PASS

**Step 3: Write implementation (update all commands to use new config path)**

**File:** `cmd/upload.go`

```go
package cmd

import (
	"fmt"
	"os"

	"obsput/pkg/config"
	"obsput/pkg/obs"
	"obsput/pkg/version"

	"github.com/spf13/cobra"
)

func NewUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload binary to OBS",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			prefix, _ := cmd.Flags().GetString("prefix")

			// Check file exists
			fileInfo, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", filePath)
			}

			// Load config
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\nRun: obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			}

			// Generate version
			gen := version.NewGenerator()
			ver := gen.Generate()

			cmd.Println("Uploading:", filePath)
			cmd.Println("Version:", ver)

			// Upload to all OBS configs
			successCount := 0
			failCount := 0

			for name, obsCfg := range cfg.Configs {
				cmd.Printf("[%s]\n", name)

				client := obs.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

				result, err := client.UploadFile(filePath, ver, prefix, func(transferred int64) {})
				if err != nil {
					cmd.Printf("✗ Failed: %v\n", err)
					failCount++
					continue
				}

				if result.Success {
					cmd.Printf("✓ Uploaded: %s\n", result.URL)
					cmd.Printf("MD5: %s\n", result.MD5)
					successCount++
				} else {
					cmd.Printf("✗ Failed: %s\n", result.Error)
					failCount++
				}
			}

			cmd.Println()
			cmd.Printf("%d completed, %d failed\n", successCount, failCount)

			return nil
		},
	}
	cmd.Flags().String("prefix", "", "Path prefix for upload")
	return cmd
}
```

**File:** `cmd/list.go`

```go
package cmd

import (
	"fmt"

	"obsput/pkg/config"
	"obsput/pkg/obs"
	"obsput/pkg/output"

	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uploaded versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, _ := cmd.Flags().GetString("output")

			// Load config
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\nRun: obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			}

			formatter := output.NewFormatter()

			for name, obsCfg := range cfg.Configs {
				cmd.Printf("[%s]\n", name)

				client := obs.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)
				versions, err := client.ListVersions("")
				if err != nil {
					cmd.Printf("Error: %v\n", err)
					continue
				}

				items := make([]output.VersionItem, 0, len(versions))
				for _, v := range versions {
					items = append(items, output.VersionItem{
						Version: v.Version,
						Size:    v.Size,
						Date:    v.Date,
						Commit:  v.Commit,
						URL:     v.URL,
					})
				}

				if outputFormat == "json" {
					formatter.PrintJSON(items)
				} else {
					formatter.PrintVersionTable(items)
				}
			}

			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "table", "Output format (table/json)")
	return cmd
}
```

**File:** `cmd/delete.go`

```go
package cmd

import (
	"fmt"

	"obsput/pkg/config"
	"obsput/pkg/obs"

	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <version>",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			targetName, _ := cmd.Flags().GetString("name")

			// Load config
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\nRun: obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			}

			for name, obsCfg := range cfg.Configs {
				// If name specified, only delete from that OBS
				if targetName != "" && name != targetName {
					continue
				}

				cmd.Printf("[%s] Deleting %s...\n", name, version)

				client := obs.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)
				result := client.DeleteVersion(version)

				if result.Success {
					cmd.Printf("✓ Deleted: %s\n", version)
				} else {
					cmd.Printf("✗ Failed: %s\n", result.Error)
				}
			}

			return nil
		},
	}
	cmd.Flags().String("name", "", "OBS name to delete from")
	return cmd
}
```

**File:** `cmd/download.go`

```go
package cmd

import (
	"fmt"

	"obsput/pkg/config"
	"obsput/pkg/obs"

	"github.com/spf13/cobra"
)

func NewDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <version>",
		Short: "Show download commands for a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			outputPath, _ := cmd.Flags().GetString("output")

			// Load config
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\nRun: obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			}

			for name, obsCfg := range cfg.Configs {
				cmd.Printf("[%s]\n", name)

				client := obs.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

				// Find the version
				versions, err := client.ListVersions("")
				if err != nil {
					cmd.Printf("Error: %v\n", err)
					continue
				}

				for _, v := range versions {
					if v.Version == version {
						cmd.Printf("Version: %s\n", v.Version)
						cmd.Printf("URL: %s\n", v.URL)
						cmd.Println()
						cmd.Println("Commands:")
						cmd.Printf("  curl -O %s\n", v.URL)
						cmd.Printf("  wget %s\n", v.URL)
						if outputPath != "" {
							cmd.Printf("  curl -o %s %s\n", outputPath, v.URL)
						}
					}
				}
			}

			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output file path")
	return cmd
}
```

**Step 4: Run tests**

Run: `go test ./... -v`
Expected: All tests pass

**Step 5: Commit**

```bash
git add cmd/upload.go cmd/list.go cmd/delete.go cmd/download.go
git commit -m "feat: update all commands with new config path"
```

---

## Task 6: Final Verification

**Step 1: Build binary**

Run: `go build -o obsput main.go && ./obsput --help`
Expected: All commands work

**Step 2: Test config path**

Run: `./obsput obs init`
Expected: Creates .obsput/obsput.yaml in binary directory

**Step 3: Test empty config message**

Run: `./obsput upload ./test.bin`
Expected: Shows message about adding OBS config

**Step 4: Commit and tag**

```bash
git tag v0.3.0
git commit --allow-empty -m "chore: release v0.3.0"
```

---

## Plan Complete

**Total Tasks:** 6

**Goal achieved:**
- Config stored at {binary_dir}/.obsput/obsput.yaml
- Auto-generate template on first run
- User-friendly error messages for missing config
- All commands updated to use new config path

---

**Plan complete and saved to `docs/plans/obsput-config-path-plan.md`**

**Execution option:**
- Use superpowers:subagent-driven-development to execute task-by-task with code review
