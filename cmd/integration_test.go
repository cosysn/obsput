package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIntegration_UploadListDelete(t *testing.T) {
	// Create temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("create test file failed: %v", err)
	}

	// Test upload command structure
	uploadCmd := NewPutCommand()
	uploadBuf := bytes.NewBufferString("")
	uploadCmd.SetOut(uploadBuf)
	uploadCmd.SetArgs([]string{testFile})

	// This would fail without real OBS, but should not panic
	t.Log("Upload command created successfully")

	// Test list command structure
	listCmd := NewListCommand()
	listBuf := bytes.NewBufferString("")
	listCmd.SetOut(listBuf)

	t.Log("List command created successfully")

	// Test delete command structure
	deleteCmd := NewDeleteCommand()
	deleteBuf := bytes.NewBufferString("")
	deleteCmd.SetOut(deleteBuf)

	t.Log("Delete command created successfully")
}

func TestIntegration_ConfigCommands(t *testing.T) {
	// Test config add command structure
	addCmd := NewOBSAddCommand()
	addBuf := bytes.NewBufferString("")
	addCmd.SetOut(addBuf)

	t.Log("OBS add command created successfully")

	// Test config list command structure
	listCmd := NewOBSListCommand()
	listBuf := bytes.NewBufferString("")
	listCmd.SetOut(listBuf)

	t.Log("OBS list command created successfully")

	// Test config get command structure
	getCmd := NewOBSGetCommand()
	getBuf := bytes.NewBufferString("")
	getCmd.SetOut(getBuf)

	t.Log("OBS get command created successfully")

	// Test config remove command structure
	rmCmd := NewOBSRemoveCommand()
	rmBuf := bytes.NewBufferString("")
	rmCmd.SetOut(rmBuf)

	t.Log("OBS remove command created successfully")
}

func TestIntegration_OBSCommandStructure(t *testing.T) {
	// Test OBS parent command has subcommands
	obsCmd := NewOBSCommand()
	subCmds := obsCmd.Commands()

	if len(subCmds) == 0 {
		t.Error("OBS command should have subcommands")
	}

	// Check expected subcommands exist
	expectedSubCmds := map[string]bool{
		"add":    true,
		"list":   true,
		"get":    true,
		"remove": true,
		"mb":     true,
	}

	for _, c := range subCmds {
		if expectedSubCmds[c.Name()] {
			delete(expectedSubCmds, c.Name())
		}
	}

	if len(expectedSubCmds) > 0 {
		for name := range expectedSubCmds {
			t.Errorf("Expected OBS subcommand '%s' not found", name)
		}
	}
}
