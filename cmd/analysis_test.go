package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/git-pkgs/git-pkgs/cmd"
)

func TestIntegrityCommand(t *testing.T) {
	t.Run("shows integrity hashes", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		addFileAndCommit(t, repoDir, "package-lock.json", packageLockJSON, "Add lockfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"integrity"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("integrity failed: %v", err)
		}

		output := stdout.String()
		// Should show packages with integrity hashes
		if !strings.Contains(output, "sha512") {
			t.Errorf("expected sha512 hash in output, got: %s", output)
		}
		if !strings.Contains(output, "express") {
			t.Errorf("expected 'express' in output, got: %s", output)
		}
	})

	t.Run("drift flag filters output", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package-lock.json", packageLockJSON, "Add lockfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"integrity", "--drift"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("integrity --drift failed: %v", err)
		}

		output := stdout.String()
		// Should report no drift in a single lockfile
		if !strings.Contains(output, "No integrity drift") && !strings.Contains(output, "drift") {
			t.Errorf("expected drift-related output, got: %s", output)
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package-lock.json", packageLockJSON, "Add lockfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"integrity", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("integrity failed: %v", err)
		}

		var result []map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if len(result) == 0 {
			t.Error("expected at least one package in JSON output")
		}

		// Validate structure
		first := result[0]
		if _, ok := first["name"]; !ok {
			t.Error("expected 'name' field in integrity JSON")
		}
		if _, ok := first["integrity"]; !ok {
			t.Error("expected 'integrity' field in integrity JSON")
		}
	})

	t.Run("stateless mode works", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package-lock.json", packageLockJSON, "Add lockfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		var stdout bytes.Buffer
		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"integrity", "--stateless"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("integrity --stateless failed: %v", err)
		}
	})
}

func TestStatsCommand(t *testing.T) {
	t.Run("shows statistics", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", `{"dependencies":{"a":"1.0.0"}}`, "Add a")
		addFileAndCommit(t, repoDir, "package.json", `{"dependencies":{"b":"1.0.0"}}`, "Add b")
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add more")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"stats"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("stats failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "Dependency Statistics") {
			t.Error("expected statistics header in output")
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"stats", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("stats failed: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		// Check for expected fields
		if _, ok := result["branch"]; !ok {
			t.Error("expected 'branch' field in JSON output")
		}
		if _, ok := result["current_deps"]; !ok {
			t.Error("expected 'current_deps' field in JSON output")
		}
		if _, ok := result["deps_by_ecosystem"]; !ok {
			t.Error("expected 'deps_by_ecosystem' field in JSON output")
		}
	})

	t.Run("by-author flag works", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"stats", "--by-author"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("stats --by-author failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "Test User") {
			t.Errorf("expected 'Test User' in by-author output, got: %s", output)
		}
	})
}

func TestSearchCommand(t *testing.T) {
	t.Run("finds packages by pattern", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"search", "express"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("search failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "express") {
			t.Error("expected to find 'express' in search results")
		}
	})

	t.Run("wildcard pattern works", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"search", "lod%"}) // SQL LIKE pattern
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("search failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "lodash") {
			t.Error("expected to find 'lodash' with wildcard pattern")
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"search", "express", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("search failed: %v", err)
		}

		var result []map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if len(result) == 0 {
			t.Error("expected at least one result in JSON output")
		}

		// Verify express was found
		foundExpress := false
		for _, item := range result {
			if name, _ := item["name"].(string); name == "express" {
				foundExpress = true
				break
			}
		}
		if !foundExpress {
			t.Error("expected 'express' in search results")
		}
	})
}

func TestTreeCommand(t *testing.T) {
	t.Run("shows dependency tree", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"tree"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("tree failed: %v", err)
		}

		output := stdout.String()
		// Should show tree structure
		if !strings.Contains(output, "package.json") {
			t.Errorf("expected 'package.json' in tree output, got: %s", output)
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"tree", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("tree failed: %v", err)
		}

		var result []map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if len(result) == 0 {
			t.Error("expected at least one manifest in tree JSON output")
		}

		// Validate tree structure
		first := result[0]
		if _, ok := first["name"]; !ok {
			t.Error("expected 'name' field in tree JSON")
		}
		if _, ok := first["type"]; !ok {
			t.Error("expected 'type' field in tree JSON")
		}
	})
}

func TestBlameCommand(t *testing.T) {
	t.Run("attributes dependencies to commits", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"blame"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("blame failed: %v", err)
		}

		output := stdout.String()
		// Should show commit attribution
		if !strings.Contains(output, "express") {
			t.Errorf("expected 'express' in blame output, got: %s", output)
		}
		// Should show the author who added the dependency
		if !strings.Contains(output, "Test User") {
			t.Errorf("expected 'Test User' in blame output, got: %s", output)
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"blame", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("blame failed: %v", err)
		}

		var result []map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if len(result) == 0 {
			t.Error("expected at least one entry in blame JSON output")
		}

		// Validate structure
		first := result[0]
		if _, ok := first["name"]; !ok {
			t.Error("expected 'name' field in blame JSON")
		}
		if _, ok := first["author_name"]; !ok {
			t.Error("expected 'author_name' field in blame JSON")
		}
		if _, ok := first["sha"]; !ok {
			t.Error("expected 'sha' field in blame JSON")
		}
	})
}

func TestWhyCommand(t *testing.T) {
	t.Run("finds when package was added", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "README.md", "# Test", "Initial")
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"why", "express"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("why failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "express") {
			t.Error("expected to find 'express' in why output")
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"why", "express", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("why failed: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		// Should contain the package name
		if name, _ := result["name"].(string); name != "express" {
			t.Errorf("expected name 'express', got %q", name)
		}
		// Should have commit info
		if _, ok := result["sha"]; !ok {
			t.Error("expected 'sha' field in why JSON")
		}
		if _, ok := result["message"]; !ok {
			t.Error("expected 'message' field in why JSON")
		}
	})
}

func TestWhereCommand(t *testing.T) {
	t.Run("finds package in manifest", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"where", "express"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("where failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "package.json") {
			t.Error("expected to find package.json in where output")
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add deps")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"where", "express", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("where failed: %v", err)
		}

		var result []map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if len(result) == 0 {
			t.Error("expected at least one location in where JSON output")
		}

		// Validate structure
		first := result[0]
		if _, ok := first["file_path"]; !ok {
			t.Error("expected 'file_path' field in where JSON")
		}
		if _, ok := first["ecosystem"]; !ok {
			t.Error("expected 'ecosystem' field in where JSON")
		}
	})
}

func TestStaleCommand(t *testing.T) {
	t.Run("finds stale dependencies", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package-lock.json", packageLockJSON, "Add lockfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"stale", "--days", "0"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("stale failed: %v", err)
		}

		// With --days 0, all lockfile deps should be stale
		output := stdout.String()
		// Should show dependencies from the lockfile as stale
		if !strings.Contains(output, "express") && !strings.Contains(output, "lodash") {
			t.Errorf("expected stale dependencies in output, got: %s", output)
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package-lock.json", packageLockJSON, "Add lockfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"stale", "--days", "0", "--format", "json"})
		rootCmd.SetOut(&stdout)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("stale failed: %v", err)
		}

		var result []map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		// With --days 0, should have stale packages
		if len(result) == 0 {
			t.Error("expected stale packages in JSON output with --days 0")
		}

		// Validate structure
		first := result[0]
		if _, ok := first["name"]; !ok {
			t.Error("expected 'name' field in stale JSON")
		}
	})
}
