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

	"github.com/little-forest/hasher/core"
	"github.com/spf13/cobra"
)

const Flag_Compare_Source = "source"
const Flag_Compare_Target = "target"

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "",
	Long:  ``,
	RunE:  statusWrapper.RunE(runCompare),
}

func init() {
	rootCmd.AddCommand(compareCmd)

	compareCmd.Flags().StringP(Flag_Compare_Source, "s", "", "source hash file")
	compareCmd.Flags().StringP(Flag_Compare_Target, "t", "", "target hash file")
}

func runCompare(cmd *cobra.Command, args []string) (int, error) {
	sourceFile, _ := cmd.Flags().GetString(Flag_Compare_Source)
	targetFile, _ := cmd.Flags().GetString(Flag_Compare_Target)

	srcHashData, err := core.LoadHashData(sourceFile)
	if err != nil {
		return 1, err
	}

	targetHashData, err := core.LoadHashData(targetFile)
	if err != nil {
		return 1, err
	}

	result, err := doCompare(srcHashData, targetHashData)
	return result, err
}

func doCompare(src *core.HashStore, target *core.HashStore) (int, error) {
	for _, hash := range src.Values() {
		sames := target.Get(hash.String())

		fmt.Printf("%s\t%d", hash.Path, len(sames))
		if len(sames) > 0 {
			for _, s := range sames {
				fmt.Printf("\t%s", s.Path)
			}
		}
		fmt.Printf("\n")
	}
	return 0, nil
}
