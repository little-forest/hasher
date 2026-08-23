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

	"github.com/little-forest/hasher/hashcore"
	"github.com/little-forest/hasher/internal/hasher"
	"github.com/spf13/cobra"
)

const Flag_listHash_Out = "out"
const Flag_listHash_UpdateHash = "update-hash"

var (
	listHashOut        string
	listHashUpdateHash bool
)

// listHashCmd represents the listHash command
var listHashCmd = &cobra.Command{
	Use:          "list-hash [-u] [-o OUT_FILE] TARGET...",
	Short:        "Output hash list in TSV format",
	Long:         ``,
	RunE:         runListHash,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(listHashCmd)

	listHashCmd.Flags().StringVarP(&listHashOut, Flag_listHash_Out, "o", "", "output file path")
	listHashCmd.Flags().BoolVarP(&listHashUpdateHash, Flag_listHash_UpdateHash, "u", false, "When the hash is NOT up-to-date. Update it.")
}

func runListHash(cmd *cobra.Command, args []string) error {
	alg := hashcore.NewDefaultHashAlg()

	return listHashAll(args, alg, listHashOut, listHashUpdateHash)
}

func listHashAll(paths []string, alg *hashcore.HashAlg, outPath string, updateHash bool) error {
	showProgress := false

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
			showProgress = true
		}
	} else {
		writer = os.Stdout
	}

	var notifier hasher.ProgressNotifier

	if showProgress {
		notifier = NewHasherProgressNotifier(1, showProgress)
	} else {
		notifier = NewStdioProgressNotifier()
	}

	err := hasher.ListHash2(paths, hashcore.NewDefaultHashAlg(), writer, notifier, showProgress, updateHash)
	return err
}
