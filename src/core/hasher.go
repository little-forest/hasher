package core

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	. "github.com/little-forest/hasher/common"
	"github.com/pkg/errors"
)

const hashBufSize = 256 * 1024

const Xattr_prefix = "user.hasher"

// File size when hash is updated
const Xattr_size = Xattr_prefix + ".size"

// File modification time when hash is updated
const Xattr_modifiedTime = Xattr_prefix + ".mtime"

// Time of hash update
const Xattr_hashCheckedTime = Xattr_prefix + ".htime"

// UpdateHashStictly updates specified file's hash value.
// If the update of an attribute fails, a warning is displayed instead of returning an error.
//
//	changed : bool
//	hash value : *Hash
//	error : error
func UpdateHash(path string, alg *HashAlg, forceUpdate bool) (bool, *Hash, error) {
	changed, hash, err := UpdateHashStrictly(path, alg, forceUpdate)
	if err != nil {
		if errors.As(err, Err_updateError) {
			// Show warning and ignore error
			ShowWarn("Failed to update attribute : %s", err.Error())
			return changed, hash, nil
		} else {
			return false, nil, err
		}
	}
	return changed, hash, err
}

// UpdateHashStictly updates specified file's hash value.
// Returns an UpdateError if the update of an attribute fails.
//
//	changed : bool
//	hash value : *Hash
//	error : error
func UpdateHashStrictly(path string, alg *HashAlg, forceUpdate bool) (bool, *Hash, error) {
	file, err := OpenFile(path)
	if err != nil {
		return false, nil, err
	}
	// nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, nil, err
	}
	size := fmt.Sprint(info.Size())
	modTime := strconv.FormatInt(info.ModTime().UnixNano(), 10)

	var changed bool
	curHash := GetXattr(file, alg.AttrName)
	if curHash != "" {
		// check if existing hash value is valid
		// If the file size and modtime have not changed, it is considered correct.
		if curSize := GetXattr(file, Xattr_size); size != curSize {
			changed = true
		} else if curMtime := GetXattr(file, Xattr_modifiedTime); modTime != curMtime {
			changed = true
		}
		if !forceUpdate && !changed {
			// update only checked time
			err := updateHashCheckedTime(file)
			if err != nil {
				err = NewUpdateError(err)
			}
			hash, _ := NewHashFromString(path, alg, curHash, info.ModTime().Unix())
			return false, hash, err
		}
	}

	// do calculate hash value
	hash, err := CalcHash(path, alg)
	if err != nil {
		return false, nil, err
	}

	// update attributes
	if err := SetXattr(file, alg.AttrName, hash.String()); err != nil {
		return true, hash, NewUpdateError(err)
	}
	if err := updateHashCheckedTime(file); err != nil {
		return true, hash, NewUpdateError(err)
	}
	if err := SetXattr(file, Xattr_size, size); err != nil {
		return true, hash, NewUpdateError(err)
	}
	if err := SetXattr(file, Xattr_modifiedTime, modTime); err != nil {
		return true, hash, NewUpdateError(err)
	}

	return true, hash, nil
}

func updateHashCheckedTime(f *os.File) error {
	htime := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	if err := SetXattr(f, Xattr_hashCheckedTime, htime); err != nil {
		return err
	}
	return nil
}

func CalcHash(path string, hashAlg *HashAlg) (*Hash, error) {
	if !hashAlg.Alg.Available() {
		return nil, fmt.Errorf("no implementation")
	}

	r, err := OpenFile(path)
	if err != nil {
		return nil, err
	}

	hash := hashAlg.Alg.New()
	if _, err := io.CopyBuffer(hash, r, make([]byte, hashBufSize)); err != nil {
		return nil, err
	}

	info, _ := os.Stat(path)

	return NewHash(path, hashAlg, hash.Sum(nil), info.ModTime().Unix()), nil
}

// Get hash value.
// This function will not check hash is updated.
// When given file's hash has not been calculated, it will return nil.
func GetHash(path string, alg *HashAlg) (*Hash, error) {
	file, err := OpenFile(path)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	curHash := GetXattr(file, alg.AttrName)
	if curHash != "" {
		hash, _ := NewHashFromString(path, alg, curHash, info.ModTime().Unix())
		return hash, nil
	} else {
		return nil, nil
	}
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
	WorkerId int
	Task     UpdateTask
	Hash     string
	Message  string
	Err      error
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

func ConcurrentUpdateHash(paths []string, alg *HashAlg, numOfWorkers int, forceUpdate bool, notifier ProgressNotifier) error {
	total, err := CountAllFiles(paths, notifier.IsVerbose())
	if err != nil {
		return err
	}

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
		isSym, err := IsSymbolicLink(p)
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
		// TODO: error check
		// nolint:staticcheck
		err = filepath.WalkDir(p, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// skip symbolic link
			isSym, err := IsSymbolicLink(p)
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

func updateHashWorker(id int, tasks <-chan UpdateTask, results chan<- UpdateResult, alg *HashAlg, forceUpdate bool, notifier ProgressNotifier) {
	for t := range tasks {
		notifier.NotifyTaskStart(id, t.Path)
		changed, hash, err := UpdateHash(t.Path, alg, forceUpdate)
		hashValue := ""
		msg := ""
		if err == nil {
			hashValue = hash.String()
			if !changed {
				msg = Mark_OK
			} else {
				msg = "[UPDATED]"
			}
		} else {
			msg = Mark_Failed
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

func ListHash(dirPaths []string, alg *HashAlg, w io.Writer, watcher ProgressNotifier, verbose bool, noCheck bool) error {
	total, err := CountAllFiles(dirPaths, watcher.IsVerbose())
	if err != nil {
		return err
	}

	watcher.SetTotal(total)
	watcher.Start()

	bw := bufio.NewWriterSize(w, 16384)
	// nolint:errcheck
	defer bw.Flush()

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

			var hash *Hash
			var changed bool
			msg := ""
			absPath, _ := filepath.Abs(path)
			if !noCheck {
				changed, hash, e = UpdateHash(absPath, alg, false)
			} else {
				hash, e = GetHash(absPath, alg)
			}
			if e != nil {
				fmt.Fprintf(os.Stderr, "Failed to update hash : %s (reason : %s)\n", absPath, e.Error())
			} else {
				if hash == nil {
					fmt.Fprintf(os.Stderr, "No hash data : %s\n", absPath)
				} else {
					fmt.Fprintf(bw, "%s\n", hash.Tsv())
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
