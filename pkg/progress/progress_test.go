package progress

import (
	"bytes"
	"testing"
	"time"
)

func TestNewProgressBar(t *testing.T) {
	pb := New(100)
	if pb == nil {
		t.Fatal("New should return non-nil")
	}
}

func TestProgressBarIncrement(t *testing.T) {
	pb := New(100)
	pb.Increment(10)

	if pb.Current() != 10 {
		t.Errorf("expected current 10, got %d", pb.Current())
	}
}

func TestProgressBarFinish(t *testing.T) {
	pb := New(100)
	pb.SetTotal(100)
	pb.Increment(100)
	pb.Finish()

	if !pb.IsFinished() {
		t.Error("progress should be finished")
	}
}

func TestProgressBarRender(t *testing.T) {
	pb := New(100)
	pb.SetTotal(100)
	pb.Increment(50)

	buf := bytes.NewBufferString("")
	pb.SetWriter(buf)
	pb.Render()

	output := buf.String()
	if output == "" {
		t.Error("render should produce output")
	}
}

func TestProgressBarWriter(t *testing.T) {
	pb := New(100)
	buf := bytes.NewBufferString("")
	pb.SetWriter(buf)

	if pb.writer != buf {
		t.Error("writer should be set")
	}
}

func TestProgressBarCallback(t *testing.T) {
	var called bool
	var received int64

	pb := New(100)
	pb.SetCallback(func(transferred int64) {
		called = true
		received = transferred
	})

	pb.Increment(50)

	if !called {
		t.Error("callback should have been called")
	}
	if received != 50 {
		t.Errorf("expected 50, got %d", received)
	}
}

func TestProgressBarWithBytes(t *testing.T) {
	pb := New(1024 * 1024) // 1MB
	pb.SetTotal(1024 * 1024)

	// Simulate progress
	for i := 0; i < 10; i++ {
		pb.Increment(1024 * 100) // 100KB chunks
	}

	if pb.Current() != 1024*1000 {
		t.Errorf("expected current to be 1MB, got %d", pb.Current())
	}
}

func TestProgressBarSpeed(t *testing.T) {
	pb := New(1000)
	pb.SetTotal(1000)

	// Simulate time passing
	pb.SetStartTime(time.Now().Add(-time.Second))
	pb.Increment(500)

	speed := pb.GetSpeed()
	if speed <= 0 {
		t.Error("speed should be positive")
	}
}

func TestProgressBarTimeRemaining(t *testing.T) {
	pb := New(1000)
	pb.SetTotal(1000)
	pb.Increment(250)

	remaining := pb.GetTimeRemaining()
	if remaining < 0 {
		t.Error("time remaining should be positive")
	}
}
