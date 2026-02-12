package cmd

import (
	"bytes"
	"testing"
)

func TestPutWithProgressBar(t *testing.T) {
	cmd := NewPutCommand()
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)

	// Test that put command accepts progress display
	// For now, verify command structure
	if cmd.Use != "put <file>" {
		t.Errorf("expected 'put <file>', got '%s'", cmd.Use)
	}
}
