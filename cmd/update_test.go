package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Regression test for the bug where update without -r reported success even
// though every path had failed.
func TestUpdateReturnsErrorWhenPathIsMissing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-file")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"update", missing})
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	err := Execute(context.Background())

	assert.Error(t, err)
	// SilenceUsage keeps the whole usage text out of the failure output.
	assert.NotContains(t, out.String(), "Usage:")
}
