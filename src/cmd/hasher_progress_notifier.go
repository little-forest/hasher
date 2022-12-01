package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"path/filepath"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/morikuni/aec"
)

type progressNotfierEvent interface {
	WorkerId() int
}

type commonEvent struct {
	workerId int
}

func (e commonEvent) WorkerId() int {
	return e.workerId
}

type startEvent struct {
	commonEvent
	TaskName string
}

type doneEvent struct {
	commonEvent
	Message string
}

type progressEvent struct {
	commonEvent
	Done  int
	Total int
}

type errorEvent struct {
	commonEvent
	Message string
}

type HasherProgressNotifier struct {
	NumOfWorkers int
	Verbose      bool
	done         int
	total        int
	messages     []string
	notifyQueue  chan progressNotfierEvent
	isClosed     bool
}

func NewHasherProgressNotifier(numOfWorkers int, verbose bool) *HasherProgressNotifier {
	return &HasherProgressNotifier{
		NumOfWorkers: numOfWorkers,
		Verbose:      verbose,
		total:        -1,
		messages:     make([]string, numOfWorkers),
		notifyQueue:  make(chan progressNotfierEvent),
		isClosed:     false,
	}
}

func (n *HasherProgressNotifier) SetTotal(total int) {
	if total >= 0 {
		n.total = total
	}
}

func (n HasherProgressNotifier) IsVerbose() bool {
	return n.Verbose
}

func (n *HasherProgressNotifier) Start() {
	go n.doStart()
}

func (n *HasherProgressNotifier) Shutdown() {
	close(n.notifyQueue)

	for true {
		if n.isClosed {
			break
		}
		time.Sleep(time.Microsecond * 10)
	}
}

func (n HasherProgressNotifier) NotifyTaskStart(workerId int, taskName string) {
	e := startEvent{
		commonEvent: commonEvent{
			workerId: workerId,
		},
		TaskName: taskName,
	}
	n.notifyQueue <- e
}

func (n HasherProgressNotifier) NotifyTaskDone(workerId int, message string) {
	e := doneEvent{
		commonEvent: commonEvent{
			workerId: workerId,
		},
		Message: message,
	}
	n.notifyQueue <- e
}

func (n HasherProgressNotifier) NotifyProgress(done int, total int) {
	e := progressEvent{
		Done:  done,
		Total: total,
	}
	n.notifyQueue <- e
}

func (n HasherProgressNotifier) NotifyError(workerId int, message string) {
	e := errorEvent{
		commonEvent: commonEvent{
			workerId: workerId,
		},
		Message: message,
	}
	n.notifyQueue <- e
}

// =================================================================================

func (n HasherProgressNotifier) prepare() {
	if !n.Verbose {
		return
	}
	row := n.NumOfWorkers
	fmt.Print(strings.Repeat("\n", row))
	fmt.Print(aec.Hide)
	fmt.Print(aec.Up(uint(row)))
}

func (n *HasherProgressNotifier) updatePrograss(done int, total int) {
	if done >= 0 {
		n.done = done
	}

	if total >= 0 {
		n.total = total
	}
}

func (n *HasherProgressNotifier) doStart() {
	n.prepare()

	for event := range n.notifyQueue {
		switch e := event.(type) {
		case startEvent:
			n.showStart(e.workerId, e.TaskName)
		case doneEvent:
			n.showDone(e.workerId, e.Message)
		case progressEvent:
			n.showProgress(e.Done, e.Total)
		case errorEvent:
			n.showError(e.Message)
		default:
			n.showError(fmt.Sprintf("<ProgressNotifier> Unknown event : %T : %v", e, e))
		}
	}

	n.tearDown()
	n.isClosed = true
}

func (n *HasherProgressNotifier) showTaskMessage(workerId int) {
	fmt.Print(aec.Down(uint(workerId)))
	fmt.Print("\x1b[0K") // delete line after cursor
	fmt.Printf("[Worker-%d] : %s", workerId, n.messages[workerId])
	if workerId > 0 {
		fmt.Print(aec.PreviousLine(uint(workerId)))
	} else {
		fmt.Print("\x1b[1G") // Move to top of current line
	}
}

func (n *HasherProgressNotifier) showStart(workerId int, path string) {
	if !n.Verbose {
		return
	}

	n.messages[workerId] = n.chopPath(path)

	n.showTaskMessage(workerId)
}

func (n *HasherProgressNotifier) showDone(workerId int, message string) {
	if !n.Verbose {
		return
	}

	n.messages[workerId] = n.messages[workerId] + " " + message

	n.showTaskMessage(workerId)
}

func (n HasherProgressNotifier) showError(msg string) {
	if n.Verbose {
		// Insert one line to bottom
		fmt.Print(aec.NextLine(uint(n.NumOfWorkers)))
		fmt.Printf("\n")
		fmt.Print(aec.PreviousLine(uint(n.NumOfWorkers + 1)))

		fmt.Print("\x1b[1L") // insert one line
	}
	// TODO: when stderr is redirected to file, display is broken
	fmt.Fprintln(os.Stderr, C_lred.Apply(msg))
}

func (n *HasherProgressNotifier) showProgress(done int, total int) {
	n.updatePrograss(done, total)

	if !n.Verbose {
		return
	}

	fmt.Print(aec.NextLine(uint(n.NumOfWorkers)))
	fmt.Printf("%d / %d", n.done, n.total)
	fmt.Print(aec.PreviousLine(uint(n.NumOfWorkers)))
}

func (n HasherProgressNotifier) tearDown() {
	if !n.Verbose {
		return
	}
	row := n.NumOfWorkers + 1
	fmt.Print(aec.Down(uint(row)))
	fmt.Println()
	fmt.Print(aec.Show)
}

func (n HasherProgressNotifier) chopPath(path string) string {
	paths := strings.Split(path, string(os.PathSeparator))
	l := len(paths)
	if l > 2 {
		return filepath.Join("...", paths[l-2], paths[l-1])
	} else {
		return path
	}
}

// check implementation
var _ core.ProgressNotifier = &HasherProgressNotifier{}
