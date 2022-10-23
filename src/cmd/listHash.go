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

	"path/filepath"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
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
	// TODO: deal verbose
	// verbose, _ := cmd.Flags().GetBool(Flag_root_Verbose)
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
				// TODO: hash 出力
				if err == nil {
					printHashAsJson(hash)
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
		panic("not implemented")
		// recursive update, directory only
		// TODO: implement
		// err := updateHashConcurrently(args, alg, forceUpdate, verbose)
		// if err != nil {
		// 	errorStatus = err
		// 	status = 1
		// }
	}
	return status, errorStatus
}

func printHashAsJson(hash *core.Hash) {
	fmt.Printf("%s\n", hash.DollyTsv())
}
