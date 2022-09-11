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
	"strconv"

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

	alg := NewDefaultHashAlg(Xattr_prefix)

	status := 0
	var errorStatus error
	for _, p := range args {
		isDir, err := isDirectory(p)
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
			err = updateHashRecursively(p, alg, forceUpdate, verbose)
		} else {
			// update file
			changed, hash, err := updateHash(p, alg, forceUpdate)
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

func updateHash(path string, alg *HashAlg, forceUpdate bool) (bool, string, error) {
	file, err := openFile(path)
	if err != nil {
		return false, "", err
	}
	// nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, "", err
	}
	size := fmt.Sprint(info.Size())
	modTime := strconv.FormatInt(info.ModTime().UnixNano(), 10)

	var changed bool
	curHash := getXattr(file, alg.AttrName)
	if curHash != "" {
		if curSize := getXattr(file, Xattr_size); size != curSize {
			changed = true
		} else if curMtime := getXattr(file, Xattr_modifiedTime); modTime != curMtime {
			changed = true
		}
		if !forceUpdate && !changed {
			return false, curHash, nil
		}
	}

	hash, err := calcHashString(file, alg)
	if err != nil {
		return false, "", err
	}

	if err := setXattr(file, alg.AttrName, hash); err != nil {
		return true, hash, err
	}
	if err := setXattr(file, Xattr_size, size); err != nil {
		return true, hash, err
	}
	if err := setXattr(file, Xattr_modifiedTime, modTime); err != nil {
		return true, hash, err
	}

	return true, hash, nil
}

func updateHashRecursively(dirPath string, alg *HashAlg, forceUpdate bool, verbose bool) error {
	totalCount, err := countFiles(dirPath, verbose)
	if err != nil {
		return err
	}

	count := 1

	if verbose {
		hideCursor()
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
		_, _, updateErr := updateHash(path, alg, forceUpdate)
		if updateErr != nil {
			fmt.Fprintf(os.Stderr, "\nFailed to update hash : %s (reason : %s)\n", path, updateErr.Error())
		}
		count++

		return nil
	})
	if verbose {
		fmt.Printf("\n")
		showCursor()
	}

	return err
}
