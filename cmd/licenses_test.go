package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/git-pkgs/enrichment"
	"github.com/git-pkgs/git-pkgs/cmd"
	"github.com/git-pkgs/spdx"
)

// mockEnrichmentClient returns canned license data instead of calling external APIs.
type mockEnrichmentClient struct {
	packages map[string]*enrichment.PackageInfo
}

func (m *mockEnrichmentClient) BulkLookup(_ context.Context, purls []string) (map[string]*enrichment.PackageInfo, error) {
	result := make(map[string]*enrichment.PackageInfo)
	for _, p := range purls {
		if pkg, ok := m.packages[p]; ok {
			result[p] = pkg
		}
	}
	return result, nil
}

func (m *mockEnrichmentClient) GetVersions(_ context.Context, _ string) ([]enrichment.VersionInfo, error) {
	return nil, nil
}

func (m *mockEnrichmentClient) GetVersion(_ context.Context, _ string) (*enrichment.VersionInfo, error) {
	return nil, nil
}

// setMockEnrichment replaces the enrichment client constructor with one that
// returns a mock, and returns a cleanup function to restore the original.
func setMockEnrichment(packages map[string]*enrichment.PackageInfo) func() {
	orig := cmd.NewEnrichmentClient
	cmd.NewEnrichmentClient = func() (enrichment.Client, error) {
		return &mockEnrichmentClient{packages: packages}, nil
	}
	return func() { cmd.NewEnrichmentClient = orig }
}

// Gemfile with a known copyleft dependency (sidekiq uses LGPL)
const gemfileWithLGPL = `source 'https://rubygems.org'
gem 'sidekiq'
gem 'rails'
`

func TestLicensesCommand(t *testing.T) {
	t.Run("permissive flag detects non-permissive licenses", func(t *testing.T) {
		restore := setMockEnrichment(map[string]*enrichment.PackageInfo{
			"pkg:gem/sidekiq": {Ecosystem: "rubygems", Name: "sidekiq", License: "LGPL-3.0-or-later"},
			"pkg:gem/rails":   {Ecosystem: "rubygems", Name: "rails", License: "MIT"},
		})
		defer restore()

		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "Gemfile", gemfileWithLGPL, "Add Gemfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout, stderr bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"licenses", "--permissive"})
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)

		err := rootCmd.Execute()
		output := stdout.String()

		if !strings.Contains(output, "LGPL") {
			t.Fatalf("expected LGPL in output, got: %s", output)
		}
		if !strings.Contains(output, "FLAGGED") {
			t.Error("LGPL license detected but not flagged as non-permissive")
		}
		if err == nil {
			t.Error("expected command to return error when violations found")
		}
	})

	t.Run("deny flag blocks specified licenses", func(t *testing.T) {
		restore := setMockEnrichment(map[string]*enrichment.PackageInfo{
			"pkg:gem/sidekiq": {Ecosystem: "rubygems", Name: "sidekiq", License: "LGPL-3.0-or-later"},
			"pkg:gem/rails":   {Ecosystem: "rubygems", Name: "rails", License: "MIT"},
		})
		defer restore()

		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "Gemfile", gemfileWithLGPL, "Add Gemfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout, stderr bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"licenses", "--deny", "LGPL"})
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)

		err := rootCmd.Execute()
		output := stdout.String()

		if !strings.Contains(output, "LGPL") {
			t.Fatalf("expected LGPL in output, got: %s", output)
		}
		if !strings.Contains(output, "FLAGGED") {
			t.Error("LGPL license detected but not flagged as denied")
		}
		if err == nil {
			t.Error("expected command to return error when denied license found")
		}
	})

	t.Run("copyleft flag detects copyleft licenses", func(t *testing.T) {
		restore := setMockEnrichment(map[string]*enrichment.PackageInfo{
			"pkg:gem/sidekiq": {Ecosystem: "rubygems", Name: "sidekiq", License: "LGPL-3.0-or-later"},
			"pkg:gem/rails":   {Ecosystem: "rubygems", Name: "rails", License: "MIT"},
		})
		defer restore()

		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "Gemfile", gemfileWithLGPL, "Add Gemfile")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout, stderr bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"licenses", "--copyleft"})
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)

		err := rootCmd.Execute()
		output := stdout.String()

		if !strings.Contains(output, "LGPL") {
			t.Fatalf("expected LGPL in output, got: %s", output)
		}
		if !strings.Contains(output, "FLAGGED") {
			t.Error("LGPL license detected but not flagged as copyleft")
		}
		if err == nil {
			t.Error("expected command to return error when copyleft license found")
		}
	})

	t.Run("json output format", func(t *testing.T) {
		restore := setMockEnrichment(map[string]*enrichment.PackageInfo{
			"pkg:npm/express": {Ecosystem: "npm", Name: "express", License: "MIT"},
			"pkg:npm/lodash":  {Ecosystem: "npm", Name: "lodash", License: "MIT"},
		})
		defer restore()

		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout, stderr bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"licenses", "--format", "json"})
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)
		_ = rootCmd.Execute()

		output := stdout.String()
		if len(output) == 0 {
			t.Fatal("expected JSON output, got empty string")
		}

		var result []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
		}

		if len(result) == 0 {
			t.Fatal("expected at least one license entry")
		}
		if _, ok := result[0]["name"]; !ok {
			t.Error("expected 'name' field in license JSON")
		}
	})

	t.Run("allow flag permits only listed licenses", func(t *testing.T) {
		restore := setMockEnrichment(map[string]*enrichment.PackageInfo{
			"pkg:npm/express": {Ecosystem: "npm", Name: "express", License: "MIT"},
			"pkg:npm/lodash":  {Ecosystem: "npm", Name: "lodash", License: "BSD-3-Clause"},
		})
		defer restore()

		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")

		cleanup := chdir(t, repoDir)
		defer cleanup()

		rootCmd := cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"init"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("init failed: %v", err)
		}

		var stdout, stderr bytes.Buffer
		rootCmd = cmd.NewRootCmd()
		rootCmd.SetArgs([]string{"licenses", "--allow", "MIT"})
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)

		err := rootCmd.Execute()
		output := stdout.String()

		if !strings.Contains(output, "FLAGGED") {
			t.Error("non-MIT license (BSD-3-Clause) should be flagged when only MIT is allowed")
		}
		if err == nil {
			t.Error("expected command to return error when non-allowed license found")
		}
	})
}

func TestSpdxPermissiveCheck(t *testing.T) {
	tests := []struct {
		license    string
		permissive bool
	}{
		{"MIT", true},
		{"Apache-2.0", true},
		{"BSD-3-Clause", true},
		{"BSD-2-Clause", true},
		{"ISC", true},
		{"GPL-3.0-only", false},
		{"GPL-2.0-only", false},
		{"LGPL-3.0-or-later", false},
		{"AGPL-3.0-only", false},
		{"MPL-2.0", false},
		{"MIT OR Apache-2.0", true},
		{"MIT OR GPL-3.0-only", false}, // has non-permissive option
	}

	for _, tt := range tests {
		t.Run(tt.license, func(t *testing.T) {
			got := spdx.IsFullyPermissive(tt.license)
			if got != tt.permissive {
				t.Errorf("IsFullyPermissive(%q) = %v, want %v", tt.license, got, tt.permissive)
			}
		})
	}
}

func TestSpdxCopyleftCheck(t *testing.T) {
	tests := []struct {
		license  string
		copyleft bool
	}{
		{"MIT", false},
		{"Apache-2.0", false},
		{"GPL-3.0-only", true},
		{"GPL-2.0-only", true},
		{"LGPL-3.0-or-later", true},
		{"AGPL-3.0-only", true},
		{"MPL-2.0", true},
		{"MIT OR GPL-3.0-only", true}, // has copyleft option
		{"MIT OR Apache-2.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.license, func(t *testing.T) {
			got := spdx.HasCopyleft(tt.license)
			if got != tt.copyleft {
				t.Errorf("HasCopyleft(%q) = %v, want %v", tt.license, got, tt.copyleft)
			}
		})
	}
}

func TestSpdxNormalization(t *testing.T) {
	tests := []struct {
		input      string
		normalized string
	}{
		{"MIT", "MIT"},
		{"MIT License", "MIT"},
		{"Apache 2", "Apache-2.0"},
		{"Apache 2.0", "Apache-2.0"},
		{"GPL v3", "GPL-3.0-or-later"},
		{"GPLv3", "GPL-3.0-or-later"},
		{"LGPL 3", "LGPL-3.0-or-later"},
		{"BSD 3-Clause", "BSD-3-Clause"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := spdx.Normalize(tt.input)
			if err != nil {
				t.Fatalf("Normalize(%q) error: %v", tt.input, err)
			}
			if got != tt.normalized {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.normalized)
			}
		})
	}
}

// TestLicensesDockerPURLCanonicalization tests that docker images show proper
// names even when the API returns a canonicalized PURL different from the input.
// For example, pkg:docker/postgres becomes pkg:docker/library%2Fpostgres in the response.
func TestLicensesDockerPURLCanonicalization(t *testing.T) {
	restore := setMockEnrichment(map[string]*enrichment.PackageInfo{
		"pkg:docker/postgres": {Ecosystem: "docker", Name: "postgres", License: "PostgreSQL"},
		"pkg:docker/redis":    {Ecosystem: "docker", Name: "redis", License: "BSD-3-Clause"},
	})
	defer restore()

	const dockerCompose = `version: '3'
services:
  db:
    image: postgres:16-alpine
  cache:
    image: redis:7-alpine
`
	repoDir := createTestRepo(t)
	addFileAndCommit(t, repoDir, "docker-compose.yml", dockerCompose, "Add docker-compose")

	cleanup := chdir(t, repoDir)
	defer cleanup()

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	var stdout, stderr bytes.Buffer
	rootCmd = cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"licenses", "--format", "json"})
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	_ = rootCmd.Execute()

	output := stdout.String()
	if output == "" {
		t.Fatal("expected JSON output, got empty string")
	}

	var results []cmd.LicenseInfo
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}

	for _, r := range results {
		if strings.Contains(r.PURL, "docker") {
			if r.Name == "" {
				t.Errorf("docker package %s has empty name", r.PURL)
			}
			if r.Ecosystem == "" {
				t.Errorf("docker package %s has empty ecosystem", r.PURL)
			}
		}
	}
}

func TestSpdxDenyListMatching(t *testing.T) {
	// Simulate the deny list logic from licenses.go
	denyList := []string{"GPL v3", "LGPL"}

	// Build deny set with normalization
	denySet := make(map[string]bool)
	for _, l := range denyList {
		if normalized, err := spdx.Normalize(l); err == nil {
			denySet[normalized] = true
		}
	}

	tests := []struct {
		license string
		denied  bool
	}{
		{"GPL-3.0-or-later", true}, // matches "GPL v3" normalized
		{"GPL-3.0-only", false},    // "GPL v3" normalizes to -or-later
		{"LGPL-3.0-or-later", true},
		{"MIT", false},
		{"Apache-2.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.license, func(t *testing.T) {
			// Check if license is in deny set (with normalization)
			inDenyList := denySet[tt.license]
			if !inDenyList {
				if normalized, err := spdx.Normalize(tt.license); err == nil {
					inDenyList = denySet[normalized]
				}
			}
			if inDenyList != tt.denied {
				t.Errorf("license %q denied = %v, want %v", tt.license, inDenyList, tt.denied)
			}
		})
	}
}
