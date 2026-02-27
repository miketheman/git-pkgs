package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/git-pkgs/git-pkgs/cmd"
)

func TestChangelogRequiresArg(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"changelog"})

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stdout)

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when no package argument provided")
	}
}

func TestChangelogAcceptsPURL(t *testing.T) {
	t.Run("invalid PURL returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cleanup := chdir(t, tmpDir)
		defer cleanup()

		_, _, err := runCmd(t, "changelog", "pkg:")
		if err == nil {
			t.Error("expected error for invalid PURL")
		}
		if !strings.Contains(err.Error(), "parsing purl") {
			t.Errorf("expected purl parse error, got: %v", err)
		}
	})
}
