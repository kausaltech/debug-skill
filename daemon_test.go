package dap

import (
	"fmt"
	"testing"
)

func TestOutputBufferBoundedAtWrite(t *testing.T) {
	d := &Daemon{}

	// Write more than maxOutputLines lines — buffer must never exceed cap
	total := maxOutputLines + 500
	for i := 0; i < total; i++ {
		d.appendOutput(fmt.Sprintf("line %d\n", i))
	}

	if len(d.outputLines) > maxOutputLines {
		t.Errorf("expected at most %d lines in buffer, got %d", maxOutputLines, len(d.outputLines))
	}

	// Last line in buffer should be the last complete line written
	lastExpected := fmt.Sprintf("line %d", total-1)
	last := d.outputLines[len(d.outputLines)-1]
	if last != lastExpected {
		t.Errorf("expected last buffered line %q, got %q", lastExpected, last)
	}
}

func TestOutputBufferUnderLimit(t *testing.T) {
	d := &Daemon{}

	for i := 0; i < 10; i++ {
		d.appendOutput(fmt.Sprintf("line %d\n", i))
	}

	// All 10 lines should be present (no trimming below cap)
	if len(d.outputLines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(d.outputLines))
	}
}

func TestHandleOutput(t *testing.T) {
	d := &Daemon{}
	d.appendOutput("hello\n")
	d.appendOutput("world\n")

	resp := d.handleOutput()

	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %q", resp.Status)
	}
	if resp.Data == nil {
		t.Fatal("expected Data to be set")
	}
	if resp.Data.Output != "hello\nworld" {
		t.Errorf("expected output %q, got %q", "hello\nworld", resp.Data.Output)
	}

	// Buffer should be cleared
	if d.outputLines != nil || d.outputPartial.Len() != 0 {
		t.Error("buffer should be cleared after handleOutput")
	}
}

func TestTempBinaryCleanup(t *testing.T) {
	// Verify cleanup function is called in stopSession
	called := false
	d := &Daemon{
		cleanupFn: func() { called = true },
	}

	d.stopSession()

	if !called {
		t.Error("cleanupFn was not called during stopSession")
	}
	if d.cleanupFn != nil {
		t.Error("cleanupFn should be nil after stopSession")
	}
}

func TestTempBinaryCleanup_NilSafe(t *testing.T) {
	// Verify stopSession with nil cleanupFn doesn't panic
	d := &Daemon{}
	d.stopSession() // should not panic
}
