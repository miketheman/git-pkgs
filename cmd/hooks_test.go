package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDoUninstallHooks_AppendedLines(t *testing.T) {
	// Simulate a hook file where git-pkgs lines were appended to an existing hook
	existingContent := `#!/bin/bash
echo "my custom hook"
do_something

# git-pkgs hook
git pkgs reindex --quiet 2>/dev/null || true
`

	dir := t.TempDir()
	hooksDir := filepath.Join(dir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	hookPath := filepath.Join(hooksDir, "post-commit")
	if err := os.WriteFile(hookPath, []byte(existingContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Save original hookNames and override for test
	origHookNames := hookNames
	hookNames = []string{"post-commit"}
	defer func() { hookNames = origHookNames }()

	rootCmd := newTestCommand()
	if err := doUninstallHooks(rootCmd, hooksDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("reading hook: %v", err)
	}

	result := string(content)
	if strings.Contains(result, "git-pkgs") {
		t.Errorf("expected git-pkgs lines to be removed, got:\n%s", result)
	}
	if strings.Contains(result, "git pkgs reindex") {
		t.Errorf("expected reindex line to be removed, got:\n%s", result)
	}
	if !strings.Contains(result, "my custom hook") {
		t.Error("expected original hook content to be preserved")
	}
	if !strings.Contains(result, "do_something") {
		t.Error("expected original hook content to be preserved")
	}
}

func TestDoUninstallHooks_BlankLineBetweenMarkers(t *testing.T) {
	// Edge case: blank line between comment and command should still remove both
	existingContent := `#!/bin/bash
echo "my custom hook"

# git-pkgs hook

git pkgs reindex --quiet 2>/dev/null || true
`

	dir := t.TempDir()
	hooksDir := filepath.Join(dir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	hookPath := filepath.Join(hooksDir, "post-commit")
	if err := os.WriteFile(hookPath, []byte(existingContent), 0755); err != nil {
		t.Fatal(err)
	}

	origHookNames := hookNames
	hookNames = []string{"post-commit"}
	defer func() { hookNames = origHookNames }()

	rootCmd := newTestCommand()
	if err := doUninstallHooks(rootCmd, hooksDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("reading hook: %v", err)
	}

	result := string(content)
	if strings.Contains(result, "git-pkgs") {
		t.Errorf("expected git-pkgs lines to be removed, got:\n%s", result)
	}
	if strings.Contains(result, "git pkgs reindex") {
		t.Errorf("expected reindex line to be removed, got:\n%s", result)
	}
}

func newTestCommand() *cobra.Command {
	return &cobra.Command{Use: "test"}
}
