package core

type ProgressWatcher interface {
	Setup()
	Progress(workerId int, done int, total int, path string)
	ShowError(msg string)
	TearDown()
}
