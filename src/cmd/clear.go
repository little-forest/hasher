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
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear hash attributes",
	Long:  ``,
	RunE:  statusWrapper.RunE(runClear),
}

func init() {
	rootCmd.AddCommand(clearCmd)
}

func runClear(cmd *cobra.Command, args []string) (int, error) {
	verbose, _ := cmd.Flags().GetBool(Flag_root_Verbose)
	recuesive, _ := cmd.Flags().GetBool(Flag_root_Recursive)

	status := 0
	var errResult error
	for _, p := range args {
		isDir, err := IsDirectory(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			continue
		}

		if isDir {
			if !recuesive {
				// skip dir
				fmt.Fprintf(os.Stderr, "Skip directory : %s\n", p)
				continue
			}
			err = clearRecursively(p, verbose)
		} else {
			err = clear(p)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			status = 1
			errResult = err
		}
	}
	return status, errResult
}

func clear(path string) error {
	file, err := OpenFile(path)
	if err != nil {
		return err
	}
	return core.ClearXattr(file)
}

func clearRecursively(dirPath string, verbose bool) error {
	totalCount, err := CountFiles(dirPath, verbose)
	if err != nil {
		return err
	}

	count := 1

	if verbose {
		HideCursor()
	}
	err = filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		if verbose {
			fmt.Printf("\x1b7\x1b[0J%d/%d %s\x1b8", count, totalCount, path)
		}
		clearErr := clear(path)
		if clearErr != nil {
			fmt.Fprintf(os.Stderr, "\nFailed to clear hash : %s (reason : %s)\n", path, clearErr.Error())
		}
		count++

		return nil
	})
	if verbose {
		fmt.Printf("\n")
		ShowCursor()
	}
	return err
}
