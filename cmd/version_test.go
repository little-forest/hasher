package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The version line is part of the CLI contract (SPEC-CLI-110); its format must
// survive refactoring untouched.
func TestVersionOutput(t *testing.T) {
	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"version"})
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	assert.NoError(t, Execute(context.Background()))
	assert.Equal(t, "hasher version dev unknown built from dev on unknown\n", out.String())
}
