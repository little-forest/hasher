package core

type ProgressWatcher interface {
	// Set total number of tasks. This is optional
	SetTotal(total int)
	IsVerbose() bool
	Setup()
	Progress(workerId int, done int, total int, path string)
	ShowError(msg string)
	TearDown()
}
