package cmd_test

import (
	"strings"
	"testing"

	"github.com/git-pkgs/git-pkgs/cmd"
)

func TestParsePackageArg(t *testing.T) {
	t.Run("plain name passes through ecosystem flag", func(t *testing.T) {
		eco, name, version, err := cmd.ParsePackageArg("lodash", "npm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "npm" {
			t.Errorf("ecosystem = %q, want %q", eco, "npm")
		}
		if name != "lodash" {
			t.Errorf("name = %q, want %q", name, "lodash")
		}
		if version != "" {
			t.Errorf("version = %q, want empty", version)
		}
	})

	t.Run("plain name with empty ecosystem flag", func(t *testing.T) {
		eco, name, version, err := cmd.ParsePackageArg("rails", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "" {
			t.Errorf("ecosystem = %q, want empty", eco)
		}
		if name != "rails" {
			t.Errorf("name = %q, want %q", name, "rails")
		}
		if version != "" {
			t.Errorf("version = %q, want empty", version)
		}
	})

	t.Run("PURL extracts ecosystem name and version", func(t *testing.T) {
		eco, name, version, err := cmd.ParsePackageArg("pkg:cargo/serde@1.0.0", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "cargo" {
			t.Errorf("ecosystem = %q, want %q", eco, "cargo")
		}
		if name != "serde" {
			t.Errorf("name = %q, want %q", name, "serde")
		}
		if version != "1.0.0" {
			t.Errorf("version = %q, want %q", version, "1.0.0")
		}
	})

	t.Run("PURL without version", func(t *testing.T) {
		eco, name, version, err := cmd.ParsePackageArg("pkg:npm/lodash", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "npm" {
			t.Errorf("ecosystem = %q, want %q", eco, "npm")
		}
		if name != "lodash" {
			t.Errorf("name = %q, want %q", name, "lodash")
		}
		if version != "" {
			t.Errorf("version = %q, want empty", version)
		}
	})

	t.Run("PURL with namespace", func(t *testing.T) {
		eco, name, version, err := cmd.ParsePackageArg("pkg:npm/%40babel/core@7.24.0", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "npm" {
			t.Errorf("ecosystem = %q, want %q", eco, "npm")
		}
		if name != "@babel/core" {
			t.Errorf("name = %q, want %q", name, "@babel/core")
		}
		if version != "7.24.0" {
			t.Errorf("version = %q, want %q", version, "7.24.0")
		}
	})

	t.Run("PURL ignores ecosystem flag", func(t *testing.T) {
		eco, name, _, err := cmd.ParsePackageArg("pkg:cargo/serde@1.0.0", "npm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "cargo" {
			t.Errorf("ecosystem = %q, want %q (flag should be ignored)", eco, "cargo")
		}
		if name != "serde" {
			t.Errorf("name = %q, want %q", name, "serde")
		}
	})

	t.Run("invalid PURL returns error", func(t *testing.T) {
		_, _, _, err := cmd.ParsePackageArg("pkg:", "")
		if err == nil {
			t.Fatal("expected error for invalid PURL")
		}
		if !strings.Contains(err.Error(), "parsing purl") {
			t.Errorf("error = %q, want it to contain 'parsing purl'", err.Error())
		}
	})

	t.Run("gem PURL maps to rubygems ecosystem", func(t *testing.T) {
		eco, name, _, err := cmd.ParsePackageArg("pkg:gem/rails@7.0.0", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "rubygems" {
			t.Errorf("ecosystem = %q, want %q", eco, "rubygems")
		}
		if name != "rails" {
			t.Errorf("name = %q, want %q", name, "rails")
		}
	})

	t.Run("golang PURL with namespace", func(t *testing.T) {
		eco, name, version, err := cmd.ParsePackageArg("pkg:golang/github.com/spf13/cobra@1.8.0", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "golang" {
			t.Errorf("ecosystem = %q, want %q", eco, "golang")
		}
		if name != "github.com/spf13/cobra" {
			t.Errorf("name = %q, want %q", name, "github.com/spf13/cobra")
		}
		if version != "1.8.0" {
			t.Errorf("version = %q, want %q", version, "1.8.0")
		}
	})
}

func TestWhyAcceptsPURL(t *testing.T) {
	repoDir := createTestRepo(t)
	addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
	cleanup := chdir(t, repoDir)
	defer cleanup()

	_, _, err := runCmd(t, "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	t.Run("accepts PURL argument", func(t *testing.T) {
		stdout, _, err := runCmd(t, "why", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("why with PURL failed: %v", err)
		}
		if !strings.Contains(stdout, "lodash") {
			t.Errorf("expected 'lodash' in output, got: %s", stdout)
		}
	})

	t.Run("PURL produces same result as ecosystem flag", func(t *testing.T) {
		purlOut, _, err := runCmd(t, "why", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("why with PURL failed: %v", err)
		}
		flagOut, _, err := runCmd(t, "why", "lodash", "-e", "npm")
		if err != nil {
			t.Fatalf("why with flag failed: %v", err)
		}
		if purlOut != flagOut {
			t.Errorf("PURL output differs from flag output.\nPURL: %s\nFlag: %s", purlOut, flagOut)
		}
	})

	t.Run("invalid PURL returns error", func(t *testing.T) {
		_, _, err := runCmd(t, "why", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
	})
}

func TestHistoryAcceptsPURL(t *testing.T) {
	repoDir := createTestRepo(t)
	addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
	cleanup := chdir(t, repoDir)
	defer cleanup()

	_, _, err := runCmd(t, "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	t.Run("accepts PURL argument", func(t *testing.T) {
		stdout, _, err := runCmd(t, "history", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("history with PURL failed: %v", err)
		}
		if !strings.Contains(stdout, "lodash") {
			t.Errorf("expected 'lodash' in output, got: %s", stdout)
		}
	})

	t.Run("PURL produces same result as ecosystem flag", func(t *testing.T) {
		purlOut, _, err := runCmd(t, "history", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("history with PURL failed: %v", err)
		}
		flagOut, _, err := runCmd(t, "history", "lodash", "-e", "npm")
		if err != nil {
			t.Fatalf("history with flag failed: %v", err)
		}
		if purlOut != flagOut {
			t.Errorf("PURL output differs from flag output.\nPURL: %s\nFlag: %s", purlOut, flagOut)
		}
	})

	t.Run("invalid PURL returns error", func(t *testing.T) {
		_, _, err := runCmd(t, "history", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
	})
}

func TestWhereAcceptsPURL(t *testing.T) {
	repoDir := createTestRepo(t)
	addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
	cleanup := chdir(t, repoDir)
	defer cleanup()

	t.Run("accepts PURL argument", func(t *testing.T) {
		stdout, _, err := runCmd(t, "where", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("where with PURL failed: %v", err)
		}
		if !strings.Contains(stdout, "lodash") {
			t.Errorf("expected 'lodash' in output, got: %s", stdout)
		}
	})

	t.Run("PURL produces same result as ecosystem flag", func(t *testing.T) {
		purlOut, _, err := runCmd(t, "where", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("where with PURL failed: %v", err)
		}
		flagOut, _, err := runCmd(t, "where", "lodash", "-e", "npm")
		if err != nil {
			t.Fatalf("where with flag failed: %v", err)
		}
		if purlOut != flagOut {
			t.Errorf("PURL output differs from flag output.\nPURL: %s\nFlag: %s", purlOut, flagOut)
		}
	})

	t.Run("invalid PURL returns error", func(t *testing.T) {
		_, _, err := runCmd(t, "where", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
	})
}

func TestAddAcceptsPURL(t *testing.T) {
	t.Run("invalid PURL returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cleanup := chdir(t, tmpDir)
		defer cleanup()

		_, _, err := runCmd(t, "add", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
	})
}

func TestRemoveAcceptsPURL(t *testing.T) {
	t.Run("invalid PURL returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cleanup := chdir(t, tmpDir)
		defer cleanup()

		_, _, err := runCmd(t, "remove", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
	})
}

func TestUpdateAcceptsPURL(t *testing.T) {
	t.Run("invalid PURL returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cleanup := chdir(t, tmpDir)
		defer cleanup()

		_, _, err := runCmd(t, "update", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
	})
}

func TestBrowseAcceptsPURL(t *testing.T) {
	t.Run("invalid PURL returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cleanup := chdir(t, tmpDir)
		defer cleanup()

		_, _, err := runCmd(t, "browse", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
	})
}
