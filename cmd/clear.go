/*
Copyright © 2022 Yusuke KOMORI

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/little-forest/hasher/hashcore"
	"github.com/little-forest/hasher/internal/fsutil"
	"github.com/little-forest/hasher/internal/hasher"
	"github.com/little-forest/hasher/internal/term"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:          "clear",
	Short:        "Clear hash attributes",
	Long:         ``,
	RunE:         runClear,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(clearCmd)
}

func runClear(cmd *cobra.Command, args []string) error {
	var errResult error
	for _, p := range args {
		ftype, err := hashcore.CheckFileType(p)
		if err != nil {
			term.ShowError(err)
			continue
		}

		if ftype == hashcore.SymbolicLink {
			// skip symlink
			continue
		}

		if ftype == hashcore.Directory {
			if !recursive {
				// skip dir
				continue
			}
			err = clearRecursively(p, verbose)
		} else {
			err = clear(p)
		}
		if err != nil {
			term.ShowError(err)
			errResult = err
		}
	}
	return errResult
}

func clear(path string) error {
	file, err := hashcore.OpenFile(path)
	if err != nil {
		return err
	}
	// skip symlink
	if yes, _ := hashcore.IsSymbolicLink(path); yes {
		return nil
	}
	return hashcore.ClearXattr(file)
}

func clearRecursively(dirPath string, verbose bool) error {
	paths := []string{dirPath}
	totalCount := fsutil.CountAllFiles(paths, verbose)

	var n hasher.ProgressNotifier = NewHasherProgressNotifier(1, verbose)
	n.SetTotal(totalCount)
	n.Start()

	count := 0

	err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		n.NotifyTaskStart(0, path)
		resultMsg := term.Mark_OK
		clearErr := clear(path)
		if clearErr != nil {
			msg := fmt.Sprintf("Failed to clear hash : %s", clearErr.Error())
			n.NotifyError(0, msg)
			resultMsg = term.Mark_Failed
		}
		count++
		n.NotifyTaskDone(0, resultMsg)
		n.NotifyProgress(count, totalCount)

		return nil
	})
	n.Shutdown()
	return err
}
