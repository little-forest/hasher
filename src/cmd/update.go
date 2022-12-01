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
	"os"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
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
				// update file
				changed, hash, err := core.UpdateHash(p, alg, forceUpdate)
				if err == nil && verbose {
					mark := ""
					if changed {
						mark = "*"
					}
					fmt.Fprintf(os.Stdout, "%s  %s %s\n", p, hash.String(), mark)
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
		// recursive update, directory only
		err := updateHashConcurrently(args, alg, forceUpdate, verbose)
		if err != nil {
			errorStatus = err
			status = 1
		}
	}
	return status, errorStatus
}

func updateHashConcurrently(dirPaths []string, alg *core.HashAlg, forceUpdate bool, verbose bool) error {
	numOfWorkers := 1
	notifier := NewHasherProgressNotifier(numOfWorkers, verbose)

	paths := make([]string, 0)
	for _, p := range dirPaths {
		if isDir, err := IsDirectory(p); isDir && err == nil {
			paths = append(paths, p)
		} else {
			if err == nil {
				fmt.Fprintf(os.Stderr, "Not a directory, skip. : %s\n", p)
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			}
		}
	}

	if len(paths) > 0 {
		err := core.ConcurrentUpdateHash(paths, alg, numOfWorkers, forceUpdate, notifier)
		return err
	} else {
		return nil
	}
}
