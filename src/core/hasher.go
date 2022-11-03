package core

import (
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

// Update specified file's hash value
//
//	changed : bool
//	hash value : *Hash
//	error : error
func UpdateHash2(path string, alg *HashAlg, forceUpdate bool) (bool, *Hash, error) {
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
		return true, hash, err
	}
	if err := updateHashCheckedTime(file); err != nil {
		return true, hash, err
	}
	if err := SetXattr(file, Xattr_size, size); err != nil {
		return true, hash, err
	}
	if err := SetXattr(file, Xattr_modifiedTime, modTime); err != nil {
		return true, hash, err
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
	Err      error
}

func NewUpdateResult(workerId int, task UpdateTask, hash string, err error) UpdateResult {
	return UpdateResult{
		WorkerId: workerId,
		Task:     task,
		Hash:     hash,
		Err:      err,
	}
}

func ConcurrentUpdateHash(paths []string, alg *HashAlg, numOfWorkers int, forceUpdate bool, watcher ProgressWatcher) error {
	total, err := CountAllFiles(paths, watcher.IsVerbose())
	if err != nil {
		return err
	}
	watcher.SetTotal(total)

	numOfWorkers = adjustNumOfWorkers(numOfWorkers, runtime.NumCPU())

	tasks := make(chan UpdateTask, numOfWorkers*3)
	results := make(chan UpdateResult)

	// run workers
	for i := 0; i < numOfWorkers; i++ {
		go updateHashWorker(i, tasks, results, alg, forceUpdate)
	}

	watcher.Setup()

	// collect target files
	inputDone := make(chan int)
	go listTargetFiles(paths, tasks, inputDone)

	// wait
	remains := -1
	done := 0
	for {
		select {
		case r := <-results:
			done++
			if r.Err != nil {
				watcher.ShowError(r.Err.Error())
			}
			watcher.Progress(r.WorkerId, done, remains, r.Task.Path)
		case taskNum := <-inputDone:
			remains = taskNum
		}
		if remains >= 0 && done >= remains {
			break
		}
	}

	watcher.TearDown()

	return nil
}

func listTargetFiles(paths []string, tasks chan<- UpdateTask, inputDone chan<- int) {
	var numFiles int

	for _, p := range paths {
		s, err := os.Stat(p)
		if err != nil {
			// TODO: error handling
			continue
		}

		if !s.IsDir() {
			// TODO: skip symlink
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
			if !info.IsDir() {
				tasks <- NewUpdateTask(path)
				numFiles++
			}
			return nil
		})
	}

	inputDone <- numFiles
}

func updateHashWorker(id int, tasks <-chan UpdateTask, results chan<- UpdateResult, alg *HashAlg, forceUpdate bool) {
	for t := range tasks {
		_, hash, err := UpdateHash2(t.Path, alg, forceUpdate)
		hashValue := ""
		if err == nil {
			hashValue = hash.String()
		}
		results <- NewUpdateResult(id, t, hashValue, err)
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

func ListHash(dirPaths []string, alg *HashAlg, w io.Writer, watcher ProgressWatcher, verbose bool, noCheck bool) error {
	total, err := CountAllFiles(dirPaths, watcher.IsVerbose())
	if err != nil {
		return err
	}

	watcher.SetTotal(total)
	watcher.Setup()

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
				watcher.Progress(0, count, total, path)
			}

			var hash *Hash
			absPath, _ := filepath.Abs(path)
			if !noCheck {
				_, hash, e = UpdateHash2(absPath, alg, false)
			} else {
				hash, e = GetHash(absPath, alg)
			}
			if e != nil {
				fmt.Fprintf(os.Stderr, "Failed to update hash : %s (reason : %s)\n", absPath, e.Error())
			} else {
				if hash == nil {
					fmt.Fprintf(os.Stderr, "No hash data : %s\n", absPath)
				} else {
					fmt.Fprintf(w, "%s\n", hash.DollyTsv())
				}
			}
			count++
			return nil
		})
		if err != nil {
			break
		}
	}

	watcher.TearDown()

	return err
}
