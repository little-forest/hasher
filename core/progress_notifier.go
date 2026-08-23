package core

type ProgressNotifier interface {
	SetTotal(total int)
	Start()
	Shutdown()
	NotifyTaskStart(workerId int, taskName string)
	NotifyTaskDone(workerId int, message string)
	NotifyProgress(done int, total int)
	NotifyWarning(workerId int, message string)
	NotifyError(workerId int, message string)
	IsVerbose() bool
}
