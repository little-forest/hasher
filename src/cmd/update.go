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

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const Flag_Update_ForceUpdate = "force-update"

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Calculate file hash and save to extended attribute",
	Long:  ``,
	RunE:  statusWrapper.RunE(runUpdateHash),
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolP(Flag_Update_ForceUpdate, "f", false, "Force update")
}

func runUpdateHash(cmd *cobra.Command, args []string) (int, error) {
	forceUpdate, _ := cmd.Flags().GetBool(Flag_Update_ForceUpdate)
	verbose, _ := cmd.Flags().GetBool(Flag_root_Verbose)
	recuesive, _ := cmd.Flags().GetBool(Flag_root_Recursive)

	alg := core.NewDefaultHashAlg()

	status := 0
	var errorStatus error
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
			// update directory
			// err = updateHashRecursively(p, alg, forceUpdate, verbose)
			err = updateHashConcurrently(p, alg, forceUpdate, verbose)
		} else {
			// update file
			changed, hash, err := core.UpdateHash(p, alg, forceUpdate)
			if err == nil && verbose {
				mark := ""
				if changed {
					mark = "*"
				}
				fmt.Fprintf(os.Stdout, "%s  %s %s\n", p, hash, mark)
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			errorStatus = err
			status = 1
			continue
		}
	}
	return status, errorStatus
}

func updateHashRecursively(dirPath string, alg *core.HashAlg, forceUpdate bool, verbose bool) error {
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
		_, _, updateErr := core.UpdateHash(path, alg, forceUpdate)
		if updateErr != nil {
			fmt.Fprintf(os.Stderr, "\nFailed to update hash : %s (reason : %s)\n", path, updateErr.Error())
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

func updateHashConcurrently(dirPath string, alg *core.HashAlg, forceUpdate bool, verbose bool) error {
	numOfWorkers := 2
	viewer := NewHasherProgressViewer(numOfWorkers, verbose)

	fmt.Printf("Concurrent update!\n")
	paths := make([]string, 1)
	paths = append(paths, dirPath)
	err := core.ConcurrentUpdateHash(paths, alg, numOfWorkers, forceUpdate, viewer)
	return err
}
