package log_test

import (
	"bytes"
	"testing"

	"github.com/cszatmary/dot/internal/log"
)

func TestLoggerPrintf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf)
	logger.Printf("number: %d", 10)
	got := buf.String()
	want := "number: 10\n"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestLoggerNoDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf)
	logger.Debugf("number: %d", 10)
	got := buf.String()
	if got != "" {
		t.Errorf("got %s, want empty string", got)
	}
}

func TestLoggerDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf)
	logger.SetDebug(true)
	logger.Debugf("number: %d", 10)
	got := buf.String()
	want := "DEBUG: number: 10\n"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
