package log

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger, err := NewLogger("debug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewLoggerInvalidLevel(t *testing.T) {
	_, err := NewLogger("invalid")
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}
