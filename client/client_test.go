package client_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cszatmary/dot/client"
)

func TestSetupAndApply(t *testing.T) {
	homeDir := t.TempDir()
	err := os.WriteFile(
		filepath.Join(homeDir, ".gitconfig"),
		[]byte(`[pull]
	ff = only
	rebase = true
`),
		0o755,
	)
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	err = os.WriteFile(
		filepath.Join(homeDir, ".zshrc"),
		[]byte(`export PATH="/usr/local/bin:$PATH"`),
		0o755,
	)
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}

	dotClient, err := client.New(client.WithHomeDir(homeDir))
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if dotClient.IsSetup() {
		t.Error("want dot to not be setup, but it is")
	}

	err = dotClient.Setup("testdata/registry-1", false)
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if !dotClient.IsSetup() {
		t.Error("want dot to be setup, but it isn't")
	}

	err = dotClient.Apply(false)
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	filesEqual(t, filepath.Join(homeDir, ".gitconfig"), "testdata/registry-1/git/gitconfig")
	filesEqual(t, filepath.Join(homeDir, ".zshrc"), "testdata/registry-1/zsh/zshrc")
}

func filesEqual(t *testing.T, gotPath, wantPath string) {
	gotData, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("failed to read file: %s: %v", gotPath, err)
	}
	wantData, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("failed to read file: %s: %v", wantPath, err)
	}
	if !bytes.Equal(gotData, wantData) {
		t.Errorf("files not equal: got %q, want %q", gotData, wantData)
	}
}
