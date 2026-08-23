package cmd

import (
	"github.com/spf13/cobra"
)

type CobraCommandStatusWrapper struct {
	Status int
}

func (w *CobraCommandStatusWrapper) RunE(f func(cmd *cobra.Command, args []string) (int, error)) func(cmd *cobra.Command, arg []string) error {
	return func(cmd *cobra.Command, arg []string) error {
		status, err := f(cmd, arg)
		w.Status = status
		return err
	}
}

var statusWrapper = &CobraCommandStatusWrapper{}
