package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/git-pkgs/managers"
	"github.com/spf13/cobra"
)

const defaultVendorTimeout = 10 * time.Minute

func addVendorCmd(parent *cobra.Command) {
	vendorCmd := &cobra.Command{
		Use:   "vendor",
		Short: "Vendor dependencies into the project",
		Long: `Vendor dependencies into a local directory using the detected package manager.
Detects the package manager from lockfiles in the current directory
and runs the appropriate vendor command.

Examples:
  git-pkgs vendor              # vendor dependencies
  git-pkgs vendor -e go        # only vendor Go ecosystem
  git-pkgs vendor -m cargo     # force cargo`,
		RunE: runVendor,
	}

	vendorCmd.Flags().StringP("manager", "m", "", "Override detected package manager (takes precedence over -e)")
	vendorCmd.Flags().StringP("ecosystem", "e", "", "Filter to specific ecosystem")
	vendorCmd.Flags().Bool("dry-run", false, "Show what would be run without executing")
	vendorCmd.Flags().StringArrayP("extra", "x", nil, "Extra arguments to pass to package manager")
	vendorCmd.Flags().DurationP("timeout", "t", defaultVendorTimeout, "Timeout for vendor operation")
	parent.AddCommand(vendorCmd)
}

func runVendor(cmd *cobra.Command, args []string) error {
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

		builtCmds, err := BuildCommands(mgr.Name, "vendor", input)
		if err != nil {
			if !quiet {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Skipping %s: vendor not supported\n", mgr.Name)
			}
			continue
		}

		if !quiet {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Detected: %s", mgr.Name)
			if mgr.Lockfile != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), " (%s)", mgr.Lockfile)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}

		if dryRun {
			for _, c := range builtCmds {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Would run: %v\n", c)
			}
			continue
		}

		if !quiet {
			for _, c := range builtCmds {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Running: %v\n", c)
			}
		}

		if err := RunManagerCommands(ctx, dir, mgr.Name, "vendor", input, cmd.OutOrStdout(), cmd.ErrOrStderr()); err != nil {
			return fmt.Errorf("%s vendor failed: %w", mgr.Name, err)
		}
	}

	return nil
}
