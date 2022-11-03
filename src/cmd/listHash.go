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
	"io"
	"os"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/spf13/cobra"
)

const Flag_ListHash_Out = "out"
const Flag_ListHash_NoCheck = "no-check"

// listHashCmd represents the listHash command
var listHashCmd = &cobra.Command{
	Use:   "list-hash [-o OUT_FILE] TARGET_DIR...",
	Short: "output hash list in json format",
	Long:  ``,
	RunE:  statusWrapper.RunE(runListHash),
}

func init() {
	rootCmd.AddCommand(listHashCmd)

	listHashCmd.Flags().StringP(Flag_ListHash_Out, "o", "", "output file path")
	listHashCmd.Flags().BoolP(Flag_ListHash_NoCheck, "n", false, "Do not check if the hash is up-to-date. This may result in fast processing but incorrect results.")
}

func runListHash(cmd *cobra.Command, args []string) (int, error) {
	out, _ := cmd.Flags().GetString(Flag_ListHash_Out)
	noCheck, _ := cmd.Flags().GetBool(Flag_ListHash_NoCheck)
	alg := core.NewDefaultHashAlg()

	err := listHashAll(args, alg, out, noCheck)
	if err != nil {
		return 1, err
	} else {
		return 0, nil
	}
}

func listHashAll(dirPaths []string, alg *core.HashAlg, outPath string, noCheck bool) error {
	// check directory
	for _, p := range dirPaths {
		if err := EnsureDirectory(p); err != nil {
			return err
		}
	}

	verbose := false
	var writer io.Writer
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		// nolint:errcheck
		defer f.Close()
		writer = f
		verbose = true
	} else {
		writer = os.Stdout
	}

	watcher := NewHasherProgressViewer(1, verbose)

	err := core.ListHash(dirPaths, core.NewDefaultHashAlg(), writer, watcher, verbose, noCheck)
	return err
}
