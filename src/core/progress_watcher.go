package core

type ProgressWatcher interface {
	// Set total number of tasks. This is optional
	SetTotal(total int)
	IsVerbose() bool
	Setup()
	TaskStart(workerId int, taskName string)
	TaskDone(workerId int, done int, total int, message string)
	ShowError(message string)
	TearDown()
}
