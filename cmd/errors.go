package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// errSilent reports failure through the exit code without printing anything.
// Commands returning it must set SilenceErrors: true.
var errSilent = errors.New("")

// printErr writes err in the same shape cobra would. Commands that set
// SilenceErrors use it so that real errors stay visible while errSilent does
// not print anything.
func printErr(cmd *cobra.Command, err error) {
	cmd.PrintErrln(cmd.ErrPrefix(), err.Error())
}

// exactArgsOrSilent validates the argument count like cobra.ExactArgs but
// prints the error itself, so commands with SilenceErrors keep reporting a
// wrong argument count.
func exactArgsOrSilent(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(n)(cmd, args); err != nil {
			printErr(cmd, err)
			return errSilent
		}
		return nil
	}
}
