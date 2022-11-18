package core

import (
	"fmt"
	"time"
)

type ProgressEvent interface {
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
	Done    int
	Total   int
	Message string
}

type errorEvent struct {
	commonEvent
	Message string
}

type ProgressNotifier struct {
	notifyQueue chan ProgressEvent
	watcher     ProgressWatcher
	isClosed    bool
}

func NewProgressNotifier(watcher ProgressWatcher) *ProgressNotifier {
	// `total` and `verbose` must already be set in the ProgressWatcher.
	n := &ProgressNotifier{
		notifyQueue: make(chan ProgressEvent),
		watcher:     watcher,
		isClosed:    false,
	}
	return n
}

func (n ProgressNotifier) NotifyTaskStart(workerId int, taskName string) {
	e := startEvent{
		commonEvent: commonEvent{
			workerId: workerId,
		},
		TaskName: taskName,
	}
	n.notifyQueue <- e
}

func (n ProgressNotifier) NotifyTaskDone(workerId int, done int, total int, message string) {
	e := doneEvent{
		commonEvent: commonEvent{
			workerId: workerId,
		},
		Done:    done,
		Total:   total,
		Message: message,
	}
	n.notifyQueue <- e
}

func (n ProgressNotifier) NotifyError(workerId int, message string) {
	e := errorEvent{
		commonEvent: commonEvent{
			workerId: workerId,
		},
		Message: message,
	}
	n.notifyQueue <- e
}

func (n *ProgressNotifier) Shutdown() {
	close(n.notifyQueue)

	for true {
		if n.isClosed {
			break
		}
		time.Sleep(time.Microsecond * 10)
	}
}

func (n *ProgressNotifier) Start() {
	go n.doStart()
}

func (n *ProgressNotifier) doStart() {
	n.watcher.Setup()

	for event := range n.notifyQueue {
		switch e := event.(type) {
		case startEvent:
			n.watcher.TaskStart(e.workerId, e.TaskName)
		case doneEvent:
			n.watcher.TaskDone(e.workerId, e.Done, e.Total, e.Message)
		case errorEvent:
			n.watcher.ShowError(e.Message)
		default:
			n.watcher.ShowError(fmt.Sprintf("<ProgressNotifier> Unknown event : %T : %v", e, e))
		}
	}

	n.watcher.TearDown()
	n.isClosed = true
}
