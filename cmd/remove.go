package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/git-pkgs/managers"
	"github.com/git-pkgs/resolve"
	_ "github.com/git-pkgs/resolve/parsers"
	"github.com/spf13/cobra"
)

const defaultRemoveTimeout = 5 * time.Minute

func addRemoveCmd(parent *cobra.Command) {
	removeCmd := &cobra.Command{
		Use:     "remove <package>",
		Aliases: []string{"rm"},
		Short:   "Remove a dependency",
		Long: `Remove a package dependency using the detected package manager.
Detects the package manager from lockfiles in the current directory
and runs the appropriate remove command.

Examples:
  git-pkgs remove lodash
  git-pkgs remove rails
  git-pkgs remove lodash -e npm`,
		Args: cobra.ExactArgs(1),
		RunE: runRemove,
	}

	removeCmd.Flags().StringP("manager", "m", "", "Override detected package manager (takes precedence over -e)")
	removeCmd.Flags().StringP("ecosystem", "e", "", "Filter to specific ecosystem")
	removeCmd.Flags().Bool("dry-run", false, "Show what would be run without executing")
	removeCmd.Flags().StringArrayP("extra", "x", nil, "Extra arguments to pass to package manager")
	removeCmd.Flags().DurationP("timeout", "t", defaultRemoveTimeout, "Timeout for remove operation")
	parent.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	managerOverride, _ := cmd.Flags().GetString("manager")
	ecosystemFlag, _ := cmd.Flags().GetString("ecosystem")

	ecosystem, pkg, _, err := ParsePackageArg(args[0], ecosystemFlag)
	if err != nil {
		return err
	}
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	quiet, _ := cmd.Flags().GetBool("quiet")
	extra, _ := cmd.Flags().GetStringArray("extra")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	dir, err := getWorkingDir()
	if err != nil {
		return err
	}

	var mgr *DetectedManager
	if managerOverride != "" {
		mgr = &DetectedManager{Name: managerOverride}
	} else {
		detected, err := DetectManagers(dir)
		if err != nil {
			return fmt.Errorf("detecting package managers: %w", err)
		}
		if ecosystem != "" {
			detected = FilterByEcosystem(detected, ecosystem)
			if len(detected) == 0 {
				return fmt.Errorf("no %s package manager detected", ecosystem)
			}
		}
		mgr, err = PromptForManager(detected, cmd.OutOrStdout(), os.Stdin)
		if err != nil {
			return err
		}
	}

	if !quiet {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Detected: %s", mgr.Name)
		if mgr.Lockfile != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), " (%s)", mgr.Lockfile)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	input := managers.CommandInput{
		Args: map[string]string{
			"package": pkg,
		},
		Extra: extra,
	}

	builtCmds, err := BuildCommands(mgr.Name, "remove", input)
	if err != nil {
		return fmt.Errorf("building command: %w", err)
	}

	if dryRun {
		for _, c := range builtCmds {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would run: %v\n", c)
		}
		return nil
	}

	// Check if the package is a transitive dependency before removing
	if err := checkTransitiveDep(dir, mgr.Name, pkg); err != nil {
		return err
	}

	if !quiet {
		for _, c := range builtCmds {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Running: %v\n", c)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := RunManagerCommands(ctx, dir, mgr.Name, "remove", input, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
		return fmt.Errorf("remove failed: %w", err)
	}

	return nil
}

// checkTransitiveDep runs the manager's resolve command and checks whether pkg
// is a direct or transitive dependency. Returns an error only when the package
// is found exclusively as a transitive dep. If resolve isn't supported or fails,
// the check is silently skipped so removal can proceed.
func checkTransitiveDep(dir, managerName, pkg string) error {
	resolveInput := managers.CommandInput{}
	_, err := BuildCommands(managerName, "resolve", resolveInput)
	if err != nil {
		return nil // resolve not supported, skip check
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultResolveTimeout)
	defer cancel()

	var stdout bytes.Buffer
	if err := RunManagerCommands(ctx, dir, managerName, "resolve", resolveInput, &stdout, &bytes.Buffer{}); err != nil {
		return nil // resolve failed, skip check
	}

	result, err := resolve.Parse(managerName, stdout.Bytes())
	if err != nil {
		return nil // parse failed, skip check
	}

	isDirect, isTransitive := findDepInTree(result.Direct, pkg)
	if !isDirect && isTransitive {
		return fmt.Errorf("cannot remove %s: it is a transitive dependency -- remove the direct dependency that requires it instead", pkg)
	}

	return nil
}

// findDepInTree searches the dependency tree for a package by name.
// Returns whether the package appears as a direct dependency and/or as a
// transitive dependency (nested under any direct dep).
func findDepInTree(direct []*resolve.Dep, name string) (isDirect bool, isTransitive bool) {
	for _, dep := range direct {
		if dep.Name == name {
			isDirect = true
		}
		if hasTransitiveDep(dep.Deps, name) {
			isTransitive = true
		}
	}
	return isDirect, isTransitive
}

func hasTransitiveDep(deps []*resolve.Dep, name string) bool {
	for _, dep := range deps {
		if dep.Name == name {
			return true
		}
		if hasTransitiveDep(dep.Deps, name) {
			return true
		}
	}
	return false
}
