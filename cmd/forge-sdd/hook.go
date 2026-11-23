package main

import (
	"github.com/spf13/cobra"

	"github.com/forge-sdd/cli/internal/hook"
)

var hookCmd = &cobra.Command{
	Use:    "hook",
	Short:  "Wrappers dos hooks do Claude Code (uso interno)",
	Hidden: true,
}

var hookPRGuardCmd = &cobra.Command{
	Use:    "pr-guard",
	Short:  "Roda scripts/pr_guard.py --hook; sem script ou sem uv, avisa e sai 0",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		if code := hook.RunPRGuard(hook.ProjectRoot(), cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()); code != 0 {
			return exitCodeErr{code: code}
		}
		return nil
	},
}

func init() {
	hookCmd.AddCommand(hookPRGuardCmd)
	rootCmd.AddCommand(hookCmd)
}
