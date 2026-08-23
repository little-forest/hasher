package hasher

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/little-forest/hasher/hashcore"
	"github.com/little-forest/hasher/internal/fsutil"
	"github.com/little-forest/hasher/internal/term"
	"github.com/pkg/errors"
)

// Err_updateError is a sentinel used with errors.As to detect an UpdateError.
var Err_updateError = &hashcore.UpdateError{}

// UpdateHashStictly updates specified file's hash value.
// If the update of an attribute fails, a warning is displayed instead of returning an error.
//
//	changed : bool
//	hash value : *hashcore.Hash
//	error : error
func UpdateHash(path string, alg *hashcore.HashAlg, forceUpdate bool) (bool, *hashcore.Hash, error) {
	changed, hash, err := hashcore.UpdateHashStrictly(path, alg, forceUpdate)
	if err != nil {
		if errors.As(err, Err_updateError) {
			// Show warning and ignore error
			term.ShowWarn("Failed to update attribute : %s", err.Error())
			return changed, hash, nil
		} else {
			return false, nil, err
		}
	}
	return changed, hash, err
}

type UpdateTask struct {
	Path string
}

func NewUpdateTask(path string) UpdateTask {
	return UpdateTask{
		Path: path,
	}
}

type UpdateResult struct {
	Err      error
	Task     UpdateTask
	Hash     string
	Message  string
	WorkerId int
}

func NewUpdateResult(workerId int, task UpdateTask, hash string, message string, err error) UpdateResult {
	return UpdateResult{
		WorkerId: workerId,
		Task:     task,
		Hash:     hash,
		Message:  message,
		Err:      err,
	}
}

func ConcurrentUpdateHash(paths []string, alg *hashcore.HashAlg, numOfWorkers int, forceUpdate bool, notifier ProgressNotifier) error {
	total := fsutil.CountAllFiles(paths, notifier.IsVerbose())

	notifier.SetTotal(total)
	notifier.Start()

	numOfWorkers = adjustNumOfWorkers(numOfWorkers, runtime.NumCPU())

	tasks := make(chan UpdateTask, numOfWorkers*3)
	results := make(chan UpdateResult)

	// run workers
	for i := 0; i < numOfWorkers; i++ {
		go updateHashWorker(i, tasks, results, alg, forceUpdate, notifier)
	}

	// collect target files
	inputDone := make(chan int)
	go listTargetFiles(paths, tasks, inputDone)

	// wait
	remains := -1
	done := 0
	for {
		select {
		case <-results:
			done++
			notifier.NotifyProgress(done, remains)
		case taskNum := <-inputDone:
			remains = taskNum
		}
		if remains >= 0 && done >= remains {
			break
		}
	}

	notifier.Shutdown()

	return nil
}

func listTargetFiles(paths []string, tasks chan<- UpdateTask, inputDone chan<- int) {
	var numFiles int

	for _, p := range paths {
		// skip symbolic link
		isSym, err := hashcore.IsSymbolicLink(p)
		if err != nil || isSym {
			continue
		}

		s, err := os.Stat(p)
		if err != nil {
			// TODO: error handling
			continue
		}

		if !s.IsDir() {
			tasks <- NewUpdateTask(p)
			numFiles++
			continue
		}

		// walk directory
		// The walk error is deliberately discarded here; there is no path for
		// reporting it back to the caller yet (see ARCHITECTURE.md B4).
		// nolint:staticcheck,ineffassign
		err = filepath.WalkDir(p, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// skip symbolic link
			isSym, err := hashcore.IsSymbolicLink(p)
			if err != nil {
				return err
			}
			// skip symbolic link
			if isSym {
				return nil
			}
			if !info.IsDir() {
				tasks <- NewUpdateTask(path)
				numFiles++
			}
			return nil
		})
	}

	inputDone <- numFiles
}

func updateHashWorker(id int, tasks <-chan UpdateTask, results chan<- UpdateResult, alg *hashcore.HashAlg, forceUpdate bool, notifier ProgressNotifier) {
	for t := range tasks {
		notifier.NotifyTaskStart(id, t.Path)
		changed, hash, err := UpdateHash(t.Path, alg, forceUpdate)
		hashValue := ""
		msg := ""
		if err == nil {
			hashValue = hash.String()
			if !changed {
				msg = term.Mark_OK
			} else {
				msg = "[UPDATED]"
			}
		} else {
			msg = term.Mark_Failed
			notifier.NotifyError(id, err.Error())
		}
		notifier.NotifyTaskDone(id, msg)
		results <- NewUpdateResult(id, t, hashValue, msg, err)
	}
}

func adjustNumOfWorkers(numOfWorkers int, numOfCPU int) int {
	if numOfWorkers < 1 {
		numOfWorkers = 1
	}
	if numOfCPU <= 2 {
		return 1
	}
	if numOfWorkers > numOfCPU-1 {
		return numOfCPU - 1
	}
	return numOfWorkers
}

func ListHash(dirPaths []string, alg *hashcore.HashAlg, w io.Writer, watcher ProgressNotifier, verbose bool, noCheck bool) error {
	total := fsutil.CountAllFiles(dirPaths, watcher.IsVerbose())

	watcher.SetTotal(total)
	watcher.Start()

	bw := bufio.NewWriterSize(w, 16384)
	// nolint:errcheck
	defer bw.Flush()

	var err error
	count := 1
	for _, dp := range dirPaths {
		err = filepath.WalkDir(dp, func(path string, info fs.DirEntry, e error) error {
			if e != nil {
				return errors.Wrap(e, "failed to filepath.Walk")
			}

			if info.IsDir() {
				return nil
			}

			if verbose {
				watcher.NotifyTaskStart(0, path)
			}

			var hash *hashcore.Hash
			var changed bool
			msg := ""
			absPath, _ := filepath.Abs(path)
			if !noCheck {
				changed, hash, e = UpdateHash(absPath, alg, false)
			} else {
				hash, e = hashcore.GetHash(absPath, alg)
			}
			if e != nil {
				fmt.Fprintf(os.Stderr, "Failed to update hash : %s (reason : %s)\n", absPath, e.Error())
			} else {
				if hash == nil {
					fmt.Fprintf(os.Stderr, "No hash data : %s\n", absPath)
				} else {
					fmt.Fprintf(bw, "%s\n", hash.Tsv()) // nolint:errcheck
				}
			}
			count++
			if verbose {
				if !changed {
					msg = "[OK]"
				} else {
					msg = "[UPDATED]"
				}
				watcher.NotifyTaskDone(0, msg)
				watcher.NotifyProgress(count, total)
			}
			return nil
		})
		if err != nil {
			break
		}
	}

	watcher.Shutdown()

	return err
}

func ListHash2(paths []string, alg *hashcore.HashAlg, w io.Writer, watcher ProgressNotifier, verbose bool, updateHash bool) error {
	if verbose {
		if watcher == nil || !updateHash {
			return fmt.Errorf("parameter integrity error (may be bug!)")
		}
	}

	var total int
	if verbose {
		total = fsutil.CountAllFiles(paths, watcher.IsVerbose())

		watcher.SetTotal(total)
		watcher.Start()
	}

	bw := bufio.NewWriterSize(w, 16384)
	// nolint:errcheck
	defer bw.Flush()

	count := 1
	for _, p := range paths {

		t, err := hashcore.CheckFileType(p)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to stat : %s", err.Error())
			if verbose {
				watcher.NotifyTaskStart(0, p)
				watcher.NotifyError(0, errMsg)
				watcher.NotifyTaskDone(0, getUpdateMessage(false, err))
				watcher.NotifyProgress(count, total)
				count++
			} else {
				term.ShowErrorMsg(errMsg)
			}
		}

		switch t {
		case hashcore.RegularFile:
			watcher.NotifyTaskStart(0, p)
			updated, err := listSingleFileHash(p, bw, updateHash, alg)
			if err != nil {
				watcher.NotifyError(0, err.Error())
			}
			watcher.NotifyTaskDone(0, getUpdateMessage(updated, err))
			watcher.NotifyProgress(count, total)
			count++
		case hashcore.Directory:
			err := fsutil.WalkDir(p, func(f *os.File) error {
				watcher.NotifyTaskStart(0, f.Name())
				updated, err := listSingleFileHash(f.Name(), bw, updateHash, alg)
				if err != nil {
					watcher.NotifyError(0, err.Error())
				}
				watcher.NotifyTaskDone(0, getUpdateMessage(updated, err))
				watcher.NotifyProgress(count, total)
				count++
				return nil
			})
			if err != nil {
				errMsg := fmt.Sprintf("Failed to walkdir : %s", err.Error())
				if verbose {
					watcher.NotifyError(0, errMsg)
				} else {
					term.ShowErrorMsg(errMsg)
				}
			}
		default:
			term.ShowWarn("Unsupported file type : %s", p)
		}
	}

	if verbose {
		watcher.Shutdown()
	}

	return nil
}

func getUpdateMessage(updated bool, err error) string {
	if err != nil {
		return term.Mark_Error
	}
	if !updated {
		return term.Mark_OK
	} else {
		return term.Mark_Updated
	}
}

// listSingleFileHash shows given file's hash value.
// path is representing a regular file path,
// When update specified true, if the hash has not been computed,
// calculate it and return true if it has been updated.
func listSingleFileHash(path string, writer *bufio.Writer, update bool, alg *hashcore.HashAlg) (bool, error) {
	var hash *hashcore.Hash
	var changed bool
	var e error
	absPath, _ := filepath.Abs(path)
	if update {
		changed, hash, e = hashcore.UpdateHashStrictly(absPath, alg, false)
		if e != nil {
			return false, fmt.Errorf("failed to update hash : %s", e.Error())
		}
	} else {
		hash, e = hashcore.GetHash(absPath, alg)
		if e != nil {
			return false, fmt.Errorf("failed to get hash : %s", e.Error())
		}
		if hash == nil {
			// no-update mode is not intended for ProgresWatcher
			term.ShowWarn("The hash value has not yet been calculated. : %s", absPath)
			return false, nil
		}
	}
	fmt.Fprintf(writer, "%s\n", hash.Tsv()) // nolint:errcheck
	return changed, nil
}
