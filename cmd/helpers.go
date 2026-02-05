package cmd

import (
	"fmt"
	"strings"

	"github.com/git-pkgs/git-pkgs/internal/database"
	"github.com/git-pkgs/git-pkgs/internal/git"
)

func openDatabase() (*git.Repository, *database.DB, error) {
	repo, err := git.OpenRepository(".")
	if err != nil {
		return nil, nil, fmt.Errorf("not in a git repository: %w", err)
	}

	dbPath := repo.DatabasePath()
	if !database.Exists(dbPath) {
		return nil, nil, fmt.Errorf("database not found. Run 'git pkgs init' first")
	}

	db, err := database.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("opening database: %w", err)
	}

	return repo, db, nil
}

func resolveBranch(db *database.DB, branchName string) (*database.BranchInfo, error) {
	if branchName != "" {
		branch, err := db.GetBranch(branchName)
		if err != nil {
			return nil, fmt.Errorf("branch %q not found: %w", branchName, err)
		}
		return branch, nil
	}
	branch, err := db.GetDefaultBranch()
	if err != nil {
		return nil, fmt.Errorf("getting branch: %w", err)
	}
	return branch, nil
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func isResolvedDependency(d database.Dependency) bool {
	return d.Requirement != "" && (d.ManifestKind == "lockfile" || d.Ecosystem == "Go")
}

func filterByEcosystem(deps []database.Dependency, ecosystem string) []database.Dependency {
	if ecosystem == "" {
		return deps
	}
	var filtered []database.Dependency
	for _, d := range deps {
		if strings.EqualFold(d.Ecosystem, ecosystem) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}
