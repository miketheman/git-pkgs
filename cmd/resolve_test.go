package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-pkgs/git-pkgs/cmd"
)

func TestResolveDryRun(t *testing.T) {
	tests := []struct {
		name           string
		lockfile       string
		lockContent    string
		expectedOutput string
	}{
		{
			name:           "npm resolve",
			lockfile:       "package-lock.json",
			lockContent:    `{"lockfileVersion": 3}`,
			expectedOutput: "[npm ls --depth Infinity --json --long]",
		},
		{
			name:           "cargo resolve",
			lockfile:       "Cargo.lock",
			lockContent:    "[[package]]",
			expectedOutput: "[cargo metadata --format-version 1]",
		},
		{
			name:           "go resolve",
			lockfile:       "go.mod",
			lockContent:    "module test",
			expectedOutput: "[go mod graph]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := os.WriteFile(filepath.Join(tmpDir, tt.lockfile), []byte(tt.lockContent), 0644); err != nil {
				t.Fatalf("failed to write lockfile: %v", err)
			}

			writeManifestForLockfile(t, tmpDir, tt.lockfile)

			cleanup := chdir(t, tmpDir)
			defer cleanup()

			rootCmd := cmd.NewRootCmd()
			rootCmd.SetArgs([]string{"resolve", "--dry-run"})

			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)

			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("resolve --dry-run failed: %v", err)
			}

			output := stdout.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.expectedOutput, output)
			}
		})
	}
}

func TestResolveSkipsUnsupported(t *testing.T) {
	tmpDir := t.TempDir()

	// brew has no resolve command
	if err := os.WriteFile(filepath.Join(tmpDir, "Brewfile"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to write Brewfile: %v", err)
	}
	// Also add npm so we have something that works
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write lockfile: %v", err)
	}

	cleanup := chdir(t, tmpDir)
	defer cleanup()

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"resolve", "--dry-run", "-m", "brew"})

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("resolve --dry-run failed: %v", err)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "resolve not supported") {
		t.Errorf("expected 'resolve not supported' in stderr, got:\n%s", errOutput)
	}
}

func TestResolveManagerOverride(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write lockfile: %v", err)
	}

	cleanup := chdir(t, tmpDir)
	defer cleanup()

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"resolve", "--dry-run", "-m", "pnpm"})

	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("resolve --dry-run failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "pnpm list") {
		t.Errorf("expected pnpm list, got:\n%s", output)
	}
}
