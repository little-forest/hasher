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
	"crypto"
	"fmt"
	"os"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/spf13/cobra"
)

const DefaultHashAlgorithm = crypto.SHA1

const Flag_Calc_NoShowPath = "no-show-path"

// calcCmd represents the calc command
var calcCmd = &cobra.Command{
	Use:   "calc",
	Short: "Calculate hash value and show",
	Long:  ``,
	RunE:  statusWrapper.RunE(runCalcHash),
}

func init() {
	rootCmd.AddCommand(calcCmd)

	calcCmd.Flags().BoolP(Flag_Calc_NoShowPath, "n", false, "don't show path")
}

func runCalcHash(cmd *cobra.Command, args []string) (int, error) {
	for _, v := range args {
		if isDir, _ := IsDirectory(v); isDir {
			// skip directory
			continue
		}
		if isSymLink, _ := IsSymbolicLink(v); isSymLink {
			// skip symlink
			continue
		}

		hash, err := core.CalcHash(v, core.NewDefaultHashAlg())
		if err != nil {
			ShowError(err)
			continue
		}
		if f, _ := cmd.Flags().GetBool(Flag_Calc_NoShowPath); !f {
			fmt.Fprintf(os.Stdout, "%s\t%s\n", hash, v)
		} else {
			fmt.Fprintf(os.Stdout, "%s\n", hash)
		}
	}
	return 0, nil
}
