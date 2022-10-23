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

// listHashCmd represents the listHash command
var listHashCmd = &cobra.Command{
	Use:   "listHash",
	Short: "output hash list in json format",
	Long:  ``,
	RunE:  statusWrapper.RunE(runListHash),
}

func init() {
	rootCmd.AddCommand(listHashCmd)
}

func runListHash(cmd *cobra.Command, args []string) (int, error) {
	verbose, _ := cmd.Flags().GetBool(Flag_root_Verbose)
	recuesive, _ := cmd.Flags().GetBool(Flag_root_Recursive)

	alg := core.NewDefaultHashAlg()

	status := 0
	var errorStatus error

	if !recuesive {
		// normal update, file only
		for _, p := range args {
			isDir, err := IsDirectory(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				continue
			}

			if isDir {
				// skip dir
				fmt.Fprintf(os.Stderr, "Skip directory : %s\n", p)
				continue
			} else {
				// list hash
				absPath, _ := filepath.Abs(p)
				_, hash, err := core.UpdateHash2(absPath, alg, false)
				if err == nil {
					printHashAsDollyFormat(hash)
				}
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				errorStatus = err
				status = 1
				continue
			}
		}
	} else {
		// recursive list, directory only
		for _, p := range args {
			isDir, err := IsDirectory(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				continue
			}

			if isDir {
				err = listHashRecursively(p, alg, verbose)
			} else {
				// skip not directory
				fmt.Fprintf(os.Stderr, "Skip non directory : %s\n", p)
				continue
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				errorStatus = err
				status = 1
				continue
			}
		}
	}
	return status, errorStatus
}

func printHashAsDollyFormat(hash *core.Hash) {
	fmt.Printf("%s\n", hash.DollyTsv())
}

func listHashRecursively(dirPath string, alg *core.HashAlg, verbose bool) error {
	err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, e error) error {
		if e != nil {
			return errors.Wrap(e, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		if verbose {
			// TODO: verbose
			// fmt.Printf("\x1b7\x1b[0J%d/%d %s\x1b8", count, totalCount, path)
			fmt.Printf("verbose\n")
		}
		absPath, _ := filepath.Abs(path)
		_, hash, e := core.UpdateHash2(absPath, alg, false)
		if e != nil {
			fmt.Fprintf(os.Stderr, "\nFailed to update hash : %s (reason : %s)\n", absPath, e.Error())
		} else {
			printHashAsDollyFormat(hash)
		}
		return nil
	})
	return err
}
