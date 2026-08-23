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

	"github.com/little-forest/hasher/hashcore"
	"github.com/little-forest/hasher/internal/hasher"
	"github.com/spf13/cobra"
)

const Flag_Update_ForceUpdate = "force-update"

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Calculate file hash and save to extended attribute",
	Long:         ``,
	RunE:         runUpdateHash,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolP(Flag_Update_ForceUpdate, "f", false, "Force update")
}

func runUpdateHash(cmd *cobra.Command, args []string) error {
	forceUpdate, _ := cmd.Flags().GetBool(Flag_Update_ForceUpdate)
	verbose, _ := cmd.Flags().GetBool(Flag_root_Verbose)
	recuesive, _ := cmd.Flags().GetBool(Flag_root_Recursive)

	alg := hashcore.NewDefaultHashAlg()

	if recuesive {
		// recursive update, directory only
		return updateHashConcurrently(args, alg, forceUpdate, verbose)
	}

	// normal update, file only
	var errResult error
	for _, p := range args {
		isDir, err := hashcore.IsDirectory(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			errResult = err
			continue
		}

		if isDir {
			// skip dir
			fmt.Fprintf(os.Stderr, "Skip directory : %s\n", p)
			continue
		}

		// update file
		changed, hash, err := hasher.UpdateHash(p, alg, forceUpdate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			errResult = err
			continue
		}

		if verbose {
			mark := ""
			if changed {
				mark = "*"
			}
			fmt.Fprintf(os.Stdout, "%s  %s %s\n", p, hash.String(), mark) // nolint:errcheck
		}
	}
	return errResult
}

func updateHashConcurrently(dirPaths []string, alg *hashcore.HashAlg, forceUpdate bool, verbose bool) error {
	numOfWorkers := 1
	notifier := NewHasherProgressNotifier(numOfWorkers, verbose)

	paths := make([]string, 0)
	for _, p := range dirPaths {
		if isDir, err := hashcore.IsDirectory(p); isDir && err == nil {
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
		err := hasher.ConcurrentUpdateHash(paths, alg, numOfWorkers, forceUpdate, notifier)
		return err
	} else {
		return nil
	}
}
