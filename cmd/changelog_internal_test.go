package cmd

import (
	"testing"
)

func TestDetectEcosystem(t *testing.T) {
	t.Run("known manager flag without detection", func(t *testing.T) {
		eco, err := detectEcosystem("cargo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "cargo" {
			t.Errorf("got %q, want %q", eco, "cargo")
		}
	})

	t.Run("npm manager flag", func(t *testing.T) {
		eco, err := detectEcosystem("npm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "npm" {
			t.Errorf("got %q, want %q", eco, "npm")
		}
	})

	t.Run("lockfile manager maps to ecosystem", func(t *testing.T) {
		eco, err := detectEcosystem("pnpm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if eco != "npm" {
			t.Errorf("got %q, want %q", eco, "npm")
		}
	})

	t.Run("unknown manager flag", func(t *testing.T) {
		_, err := detectEcosystem("nonexistent-manager")
		if err == nil {
			t.Fatal("expected error for unknown manager")
		}
	})
}
