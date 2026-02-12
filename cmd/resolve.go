package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/git-pkgs/managers"
	"github.com/spf13/cobra"
)

const defaultResolveTimeout = 5 * time.Minute

func addResolveCmd(parent *cobra.Command) {
	resolveCmd := &cobra.Command{
		Use:   "resolve",
		Short: "Print dependency graph from the local package manager",
		Long: `Run the detected package manager's dependency graph command and print
the raw output. The output format depends on the manager: some produce
JSON (npm, cargo, pip), others produce text trees (go, maven, poetry).

Assumes dependencies are already installed. Run 'git-pkgs install' first
if needed.

Examples:
  git-pkgs resolve              # resolve dependencies
  git-pkgs resolve -e go        # only resolve Go ecosystem
  git-pkgs resolve -m cargo     # force cargo`,
		RunE: runResolve,
	}

	resolveCmd.Flags().StringP("manager", "m", "", "Override detected package manager (takes precedence over -e)")
	resolveCmd.Flags().StringP("ecosystem", "e", "", "Filter to specific ecosystem")
	resolveCmd.Flags().Bool("dry-run", false, "Show what would be run without executing")
	resolveCmd.Flags().StringArrayP("extra", "x", nil, "Extra arguments to pass to package manager")
	resolveCmd.Flags().DurationP("timeout", "t", defaultResolveTimeout, "Timeout for resolve operation")
	parent.AddCommand(resolveCmd)
}

func runResolve(cmd *cobra.Command, args []string) error {
	managerOverride, _ := cmd.Flags().GetString("manager")
	ecosystem, _ := cmd.Flags().GetString("ecosystem")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	quiet, _ := cmd.Flags().GetBool("quiet")
	extra, _ := cmd.Flags().GetStringArray("extra")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	dir, err := getWorkingDir()
	if err != nil {
		return err
	}

	detected, err := DetectManagers(dir)
	if err != nil {
		return fmt.Errorf("detecting package managers: %w", err)
	}

	if len(detected) == 0 {
		return fmt.Errorf("no package manager detected in %s", dir)
	}

	if ecosystem != "" {
		detected = FilterByEcosystem(detected, ecosystem)
		if len(detected) == 0 {
			return fmt.Errorf("no %s package manager detected", ecosystem)
		}
	}

	if managerOverride != "" {
		detected = []DetectedManager{{Name: managerOverride}}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, mgr := range detected {
		input := managers.CommandInput{
			Extra: extra,
		}

		builtCmds, err := BuildCommands(mgr.Name, "resolve", input)
		if err != nil {
			if !quiet {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Skipping %s: resolve not supported\n", mgr.Name)
			}
			continue
		}

		if !quiet {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Detected: %s", mgr.Name)
			if mgr.Lockfile != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), " (%s)", mgr.Lockfile)
			}
			_, _ = fmt.Fprintln(cmd.ErrOrStderr())
		}

		if dryRun {
			for _, c := range builtCmds {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would run: %v\n", c)
			}
			continue
		}

		if !quiet {
			for _, c := range builtCmds {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Running: %v\n", c)
			}
		}

		if err := RunManagerCommands(ctx, dir, mgr.Name, "resolve", input, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
			return fmt.Errorf("%s resolve failed: %w", mgr.Name, err)
		}
	}

	return nil
}
