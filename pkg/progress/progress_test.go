package progress

import (
	"testing"
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
