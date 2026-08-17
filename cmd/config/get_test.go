package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeConfig writes a YAML config file to a temp dir and points
// os.UserHomeDir at it by monkeypatching the test's HOME env var.
func writeConfig(t *testing.T, content string) {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".ibmverify")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
}

func TestConfigGet_singleKey(t *testing.T) {
	writeConfig(t, "tenant: https://example.verify.ibm.com\n")

	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGet(getCmd, []string{"tenant"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "https://example.verify.ibm.com" {
		t.Errorf("expected tenant URL, got %q", got)
	}
}

func TestConfigGet_allKeys(t *testing.T) {
	writeConfig(t, "tenant: https://example.verify.ibm.com\nsubject: bretton\n")

	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGet(getCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "https://example.verify.ibm.com") {
		t.Errorf("expected tenant in output, got: %s", out)
	}
	if !strings.Contains(out, "bretton") {
		t.Errorf("expected subject in output, got: %s", out)
	}
}

func TestConfigGet_unknownKey(t *testing.T) {
	writeConfig(t, "tenant: https://example.verify.ibm.com\n")

	err := runGet(getCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConfigGet_missingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // empty home — no config file

	err := runGet(getCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}
	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("unexpected error: %v", err)
	}
}
