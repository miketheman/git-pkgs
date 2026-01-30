package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiff_ManifestDeleted(t *testing.T) {
	// Create a temp repo
	dir := t.TempDir()

	// Initialize git
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Create package.json with deps
	pkgJSON := `{
  "name": "test",
  "dependencies": {
    "lodash": "^4.0.0",
    "react": "^18.0.0"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	run("add", "package.json")
	run("commit", "-m", "Initial")

	// Now delete the manifest
	if err := os.Remove(filepath.Join(dir, "package.json")); err != nil {
		t.Fatal(err)
	}

	// Change to repo dir
	oldDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	// Run diff (HEAD vs working tree)
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"diff"})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&bytes.Buffer{})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("diff command failed: %v", err)
	}
	out := buf.String()

	// Should show both packages as removed
	if !containsString(out, "lodash") {
		t.Errorf("expected lodash to be shown as removed, got:\n%s", out)
	}
	if !containsString(out, "react") {
		t.Errorf("expected react to be shown as removed, got:\n%s", out)
	}
	if !containsString(out, "Removed") {
		t.Errorf("expected 'Removed' section, got:\n%s", out)
	}
}

func TestDiff_ManifestDeletedBetweenCommits(t *testing.T) {
	// Create a temp repo
	dir := t.TempDir()

	// Initialize git
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Create package.json with deps
	pkgJSON := `{
  "name": "test",
  "dependencies": {
    "lodash": "^4.0.0",
    "react": "^18.0.0"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	run("add", "package.json")
	run("commit", "-m", "Add package.json")

	// Delete manifest and commit
	if err := os.Remove(filepath.Join(dir, "package.json")); err != nil {
		t.Fatal(err)
	}
	run("add", "package.json")
	run("commit", "-m", "Remove package.json")

	// Change to repo dir
	oldDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	// Initialize database
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"init"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Run diff HEAD~1..HEAD (should show packages removed)
	rootCmd = NewRootCmd()
	rootCmd.SetArgs([]string{"diff", "HEAD~1..HEAD"})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&bytes.Buffer{})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("diff command failed: %v", err)
	}
	out := buf.String()

	// Should show both packages as removed
	if !containsString(out, "lodash") {
		t.Errorf("expected lodash to be shown as removed, got:\n%s", out)
	}
	if !containsString(out, "react") {
		t.Errorf("expected react to be shown as removed, got:\n%s", out)
	}
	if !containsString(out, "Removed") {
		t.Errorf("expected 'Removed' section, got:\n%s", out)
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
