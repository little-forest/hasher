package cmd

import (
	"fmt"
	"os"

	"github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
)

type StdioProgressNotifier struct {
}

func NewStdioProgressNotifier() *StdioProgressNotifier {
	return &StdioProgressNotifier{}
}

func (n *StdioProgressNotifier) SetTotal(total int) {
	// do nothing
}

func (n *StdioProgressNotifier) Start() {
	// do nothing
}

func (n *StdioProgressNotifier) Shutdown() {
	// do nothing
}

func (n *StdioProgressNotifier) NotifyTaskStart(workerId int, taskName string) {
	// do nothing
}

func (n *StdioProgressNotifier) NotifyTaskDone(workerId int, message string) {
	// do nothing
}

func (n *StdioProgressNotifier) NotifyProgress(done int, total int) {
	// do nothing
}

func (n *StdioProgressNotifier) NotifyWarning(workerId int, message string) {
	fmt.Fprintln(os.Stderr, common.C_yellow.Apply(message))
}

func (n *StdioProgressNotifier) NotifyError(workerId int, message string) {
	fmt.Fprintln(os.Stderr, common.C_red.Apply(message))
}

func (n *StdioProgressNotifier) IsVerbose() bool {
	return false
}

// check implementation
var _ core.ProgressNotifier = &StdioProgressNotifier{}
