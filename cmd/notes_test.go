package cmd_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/git-pkgs/git-pkgs/internal/database"
)

func TestNotesAdd(t *testing.T) {
	t.Run("adds a note", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "add", "pkg:npm/lodash@4.17.21", "-m", "approved for use", "--set", "status=approved")
		if err != nil {
			t.Fatalf("notes add failed: %v", err)
		}
		if !strings.Contains(stdout, "Added note") {
			t.Errorf("expected 'Added note' message, got: %s", stdout)
		}
	})

	t.Run("errors on duplicate without force", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "first")
		if err != nil {
			t.Fatalf("first add failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "second")
		if err == nil {
			t.Error("expected error for duplicate note")
		}
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("force overwrites existing note", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "first")
		if err != nil {
			t.Fatalf("first add failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "replaced", "--force")
		if err != nil {
			t.Fatalf("force add failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "show", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("show failed: %v", err)
		}
		if !strings.Contains(stdout, "replaced") {
			t.Errorf("expected 'replaced' message, got: %s", stdout)
		}
		if strings.Contains(stdout, "first") {
			t.Errorf("old message should not be present, got: %s", stdout)
		}
	})
}

func TestNotesShow(t *testing.T) {
	t.Run("shows a note", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "good package", "--set", "status=approved")
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "show", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("show failed: %v", err)
		}
		if !strings.Contains(stdout, "pkg:npm/lodash") {
			t.Errorf("expected purl in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "good package") {
			t.Errorf("expected message in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "status") {
			t.Errorf("expected metadata in output, got: %s", stdout)
		}
	})

	t.Run("errors for non-existent note", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "show", "pkg:npm/nonexistent")
		if err == nil {
			t.Error("expected error for non-existent note")
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "test note", "--set", "reviewer=alice")
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "show", "pkg:npm/lodash", "-f", "json")
		if err != nil {
			t.Fatalf("show json failed: %v", err)
		}

		var note database.Note
		if err := json.Unmarshal([]byte(stdout), &note); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if note.PURL != "pkg:npm/lodash" {
			t.Errorf("expected purl 'pkg:npm/lodash', got: %s", note.PURL)
		}
		if note.Message != "test note" {
			t.Errorf("expected message 'test note', got: %s", note.Message)
		}
		if note.Metadata["reviewer"] != "alice" {
			t.Errorf("expected metadata reviewer=alice, got: %v", note.Metadata)
		}
	})
}

func TestNotesList(t *testing.T) {
	t.Run("lists notes", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "note one")
		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/express", "-m", "note two")

		stdout, _, err := runCmd(t, "notes", "list")
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if !strings.Contains(stdout, "pkg:npm/lodash") {
			t.Errorf("expected lodash in list, got: %s", stdout)
		}
		if !strings.Contains(stdout, "pkg:npm/express") {
			t.Errorf("expected express in list, got: %s", stdout)
		}
	})

	t.Run("filters by namespace", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "security review", "--namespace", "security")
		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/express", "-m", "audit note", "--namespace", "audit")

		stdout, _, err := runCmd(t, "notes", "list", "--namespace", "security")
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if !strings.Contains(stdout, "pkg:npm/lodash") {
			t.Errorf("expected lodash in filtered list, got: %s", stdout)
		}
		if strings.Contains(stdout, "pkg:npm/express") {
			t.Errorf("express should not be in security namespace list, got: %s", stdout)
		}
	})

	t.Run("filters by purl substring", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "note one")
		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/express", "-m", "note two")

		stdout, _, err := runCmd(t, "notes", "list", "--purl-filter", "lodash")
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if !strings.Contains(stdout, "pkg:npm/lodash") {
			t.Errorf("expected lodash in filtered list, got: %s", stdout)
		}
		if strings.Contains(stdout, "pkg:npm/express") {
			t.Errorf("express should not be in filtered list, got: %s", stdout)
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "note one")

		stdout, _, err := runCmd(t, "notes", "list", "-f", "json")
		if err != nil {
			t.Fatalf("list json failed: %v", err)
		}

		var notes []database.Note
		if err := json.Unmarshal([]byte(stdout), &notes); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(notes))
		}
		if notes[0].PURL != "pkg:npm/lodash" {
			t.Errorf("expected purl 'pkg:npm/lodash', got: %s", notes[0].PURL)
		}
	})

	t.Run("shows no notes message", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "list")
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if !strings.Contains(stdout, "No notes found") {
			t.Errorf("expected 'No notes found' message, got: %s", stdout)
		}
	})
}

func TestNotesRemove(t *testing.T) {
	t.Run("removes a note", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "to delete")

		stdout, _, err := runCmd(t, "notes", "remove", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("remove failed: %v", err)
		}
		if !strings.Contains(stdout, "Removed") {
			t.Errorf("expected 'Removed' message, got: %s", stdout)
		}

		_, _, err = runCmd(t, "notes", "show", "pkg:npm/lodash")
		if err == nil {
			t.Error("expected error showing removed note")
		}
	})

	t.Run("errors removing non-existent note", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "remove", "pkg:npm/nonexistent")
		if err == nil {
			t.Error("expected error removing non-existent note")
		}
	})
}

func TestNotesAppend(t *testing.T) {
	t.Run("appends to existing note", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "first line", "--set", "status=pending")
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "append", "pkg:npm/lodash", "-m", "second line", "--set", "reviewer=alice")
		if err != nil {
			t.Fatalf("append failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "show", "pkg:npm/lodash", "-f", "json")
		if err != nil {
			t.Fatalf("show failed: %v", err)
		}

		var note database.Note
		if err := json.Unmarshal([]byte(stdout), &note); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if !strings.Contains(note.Message, "first line") {
			t.Errorf("expected 'first line' in message, got: %s", note.Message)
		}
		if !strings.Contains(note.Message, "second line") {
			t.Errorf("expected 'second line' in message, got: %s", note.Message)
		}
		if note.Metadata["status"] != "pending" {
			t.Errorf("expected status=pending, got: %v", note.Metadata)
		}
		if note.Metadata["reviewer"] != "alice" {
			t.Errorf("expected reviewer=alice, got: %v", note.Metadata)
		}
	})

	t.Run("creates note when none exists", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "append", "pkg:npm/lodash", "-m", "new note via append")
		if err != nil {
			t.Fatalf("append failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "show", "pkg:npm/lodash")
		if err != nil {
			t.Fatalf("show failed: %v", err)
		}
		if !strings.Contains(stdout, "new note via append") {
			t.Errorf("expected message in output, got: %s", stdout)
		}
	})
}

func TestNotesNamespaces(t *testing.T) {
	t.Run("lists namespaces with counts", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "sec note", "--namespace", "security")
		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/express", "-m", "sec note 2", "--namespace", "security")
		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "audit note", "--namespace", "audit")
		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "default note")

		stdout, _, err := runCmd(t, "notes", "namespaces")
		if err != nil {
			t.Fatalf("namespaces failed: %v", err)
		}
		if !strings.Contains(stdout, "security") {
			t.Errorf("expected 'security' in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "audit") {
			t.Errorf("expected 'audit' in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "(default)") {
			t.Errorf("expected '(default)' in output, got: %s", stdout)
		}
		if !strings.Contains(stdout, "2 notes") {
			t.Errorf("expected '2 notes' for security namespace, got: %s", stdout)
		}
	})

	t.Run("outputs json format", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "note", "--namespace", "security")
		_, _, _ = runCmd(t, "notes", "add", "pkg:npm/express", "-m", "note")

		stdout, _, err := runCmd(t, "notes", "namespaces", "-f", "json")
		if err != nil {
			t.Fatalf("namespaces json failed: %v", err)
		}

		var namespaces []database.NamespaceCount
		if err := json.Unmarshal([]byte(stdout), &namespaces); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if len(namespaces) != 2 {
			t.Fatalf("expected 2 namespaces, got %d", len(namespaces))
		}
	})

	t.Run("shows no notes message when empty", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "namespaces")
		if err != nil {
			t.Fatalf("namespaces failed: %v", err)
		}
		if !strings.Contains(stdout, "No notes found") {
			t.Errorf("expected 'No notes found', got: %s", stdout)
		}
	})
}

func TestNotesNamespace(t *testing.T) {
	t.Run("same purl different namespaces", func(t *testing.T) {
		repoDir := createTestRepo(t)
		addFileAndCommit(t, repoDir, "package.json", packageJSON, "Add package.json")
		cleanup := chdir(t, repoDir)
		defer cleanup()

		_, _, err := runCmd(t, "init")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "security note", "--namespace", "security")
		if err != nil {
			t.Fatalf("add security failed: %v", err)
		}

		_, _, err = runCmd(t, "notes", "add", "pkg:npm/lodash", "-m", "audit note", "--namespace", "audit")
		if err != nil {
			t.Fatalf("add audit failed: %v", err)
		}

		stdout, _, err := runCmd(t, "notes", "show", "pkg:npm/lodash", "--namespace", "security")
		if err != nil {
			t.Fatalf("show security failed: %v", err)
		}
		if !strings.Contains(stdout, "security note") {
			t.Errorf("expected 'security note', got: %s", stdout)
		}

		stdout, _, err = runCmd(t, "notes", "show", "pkg:npm/lodash", "--namespace", "audit")
		if err != nil {
			t.Fatalf("show audit failed: %v", err)
		}
		if !strings.Contains(stdout, "audit note") {
			t.Errorf("expected 'audit note', got: %s", stdout)
		}
	})
}
