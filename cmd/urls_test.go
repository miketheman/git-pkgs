package cmd_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/git-pkgs/git-pkgs/cmd"
)

func TestIsPURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"pkg:cargo/serde@1.0.0", true},
		{"pkg:npm/lodash", true},
		{"pkg:npm/%40scope/pkg@1.0.0", true},
		{"serde", false},
		{"lodash", false},
		{"", false},
		{"package:npm/lodash", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cmd.IsPURL(tt.input)
			if got != tt.want {
				t.Errorf("IsPURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestUrlsPURL(t *testing.T) {
	t.Run("returns urls for cargo purl", func(t *testing.T) {
		stdout, _, err := runCmd(t, "urls", "pkg:cargo/serde@1.0.0")
		if err != nil {
			t.Fatalf("urls failed: %v", err)
		}
		if !strings.Contains(stdout, "registry") {
			t.Errorf("expected 'registry' in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "purl") {
			t.Errorf("expected 'purl' in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "crates.io") {
			t.Errorf("expected 'crates.io' in output, got: %s", stdout)
		}
	})

	t.Run("returns urls for npm purl", func(t *testing.T) {
		stdout, _, err := runCmd(t, "urls", "pkg:npm/express@4.19.0")
		if err != nil {
			t.Fatalf("urls failed: %v", err)
		}
		if !strings.Contains(stdout, "npmjs") {
			t.Errorf("expected 'npmjs' in output, got: %s", stdout)
		}
	})

	t.Run("json format", func(t *testing.T) {
		stdout, _, err := runCmd(t, "urls", "pkg:cargo/serde@1.0.0", "-f", "json")
		if err != nil {
			t.Fatalf("urls json failed: %v", err)
		}

		var urls map[string]string
		if err := json.Unmarshal([]byte(stdout), &urls); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if urls["registry"] == "" {
			t.Error("expected registry URL in json output")
		}
		if urls["purl"] == "" {
			t.Error("expected purl in json output")
		}
	})

	t.Run("errors on invalid purl", func(t *testing.T) {
		_, _, err := runCmd(t, "urls", "pkg:")
		if err == nil {
			t.Error("expected error for invalid purl")
		}
	})

	t.Run("errors on unsupported ecosystem", func(t *testing.T) {
		_, _, err := runCmd(t, "urls", "pkg:nonexistent/foo@1.0.0")
		if err == nil {
			t.Error("expected error for unsupported ecosystem")
		}
		if err != nil && !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("expected 'unsupported' in error, got: %v", err)
		}
	})
}

func TestUrlsNameLookup(t *testing.T) {
	t.Run("looks up package by name", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		stdout, _, err := runCmd(t, "urls", "lodash", "-e", "npm")
		if err != nil {
			t.Fatalf("urls failed: %v", err)
		}
		if !strings.Contains(stdout, "npmjs") {
			t.Errorf("expected 'npmjs' in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "registry") {
			t.Errorf("expected 'registry' key in output, got: %s", stdout)
		}
	})

	t.Run("errors when package not found", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "urls", "nonexistent-package-xyz")
		if err == nil {
			t.Error("expected error for non-existent package")
		}
	})

	t.Run("errors when no database exists", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "README.md", "# Test", "Initial commit")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "urls", "lodash")
		if err == nil {
			t.Error("expected error without database")
		}
	})
}
