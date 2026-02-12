package cmd

import (
	"testing"
)

func TestOBSCommand(t *testing.T) {
	cmd := NewOBSCommand()
	if cmd.Use != "obs" {
		t.Errorf("expected use 'obs', got '%s'", cmd.Use)
	}
}
