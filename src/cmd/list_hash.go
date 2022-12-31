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

	"github.com/little-forest/hasher/core"
	"github.com/spf13/cobra"
)

const Flag_ListHash_Out = "out"
const Flag_ListHash_UpdateHash = "update-hash"

// listHashCmd represents the listHash command
var listHashCmd = &cobra.Command{
	Use:   "list-hash [-u] [-o OUT_FILE] TARGET...",
	Short: "Output hash list in TSV format",
	Long:  ``,
	RunE:  statusWrapper.RunE(runListHash),
}

func init() {
	rootCmd.AddCommand(listHashCmd)

	listHashCmd.Flags().StringP(Flag_ListHash_Out, "o", "", "output file path")
	listHashCmd.Flags().BoolP(Flag_ListHash_UpdateHash, "u", false, "When the hash is NOT up-to-date. Update it.")
}

func runListHash(cmd *cobra.Command, args []string) (int, error) {
	out, _ := cmd.Flags().GetString(Flag_ListHash_Out)
	updateHash, _ := cmd.Flags().GetBool(Flag_ListHash_UpdateHash)
	alg := core.NewDefaultHashAlg()

	err := listHashAll(args, alg, out, updateHash)
	if err != nil {
		return 1, err
	} else {
		return 0, nil
	}
}

func listHashAll(paths []string, alg *core.HashAlg, outPath string, updateHash bool) error {
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

		// verobse mode when the output is a file and the update flag is true
		if updateHash {
			verbose = true
		}
	} else {
		writer = os.Stdout
	}

	var notifier core.ProgressNotifier

	if verbose {
		notifier = NewHasherProgressNotifier(1, verbose)
	} else {
		notifier = NewStdioProgressNotifier()
	}

	err := core.ListHash2(paths, core.NewDefaultHashAlg(), writer, notifier, verbose, updateHash)
	return err
}
