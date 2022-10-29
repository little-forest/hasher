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
const Flag_Compare_ShowExistsOnly = "exists-only"
const Flag_Compare_ShowMissingOnly = "missing-only"
const Flag_Compare_PrintSourcePathOnly = "print-source-path-only"
const Flag_Compare_PrintZero = "print0"

const SHOW_ALWAYS = 0
const SHOW_EXISTS_ONLY = 1
const SHOW_MISSING_ONLY = 2

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare hashes",
	Long:  ``,
	RunE:  statusWrapper.RunE(runCompare),
}

func init() {
	rootCmd.AddCommand(compareCmd)

	compareCmd.Flags().StringP(Flag_Compare_Source, "s", "", "source hash file")
	compareCmd.Flags().StringP(Flag_Compare_Target, "t", "", "target hash file")
	compareCmd.Flags().BoolP(Flag_Compare_ShowExistsOnly, "e", false, "show exist files only")
	compareCmd.Flags().BoolP(Flag_Compare_ShowMissingOnly, "m", false, "show missing files only")
	compareCmd.Flags().BoolP(Flag_Compare_PrintSourcePathOnly, "f", false, "print only source file path")
	compareCmd.Flags().BoolP(Flag_Compare_PrintZero, "0", false, "separate by null character")
}

func runCompare(cmd *cobra.Command, args []string) (int, error) {
	sourceFile, _ := cmd.Flags().GetString(Flag_Compare_Source)
	targetFile, _ := cmd.Flags().GetString(Flag_Compare_Target)

	printSourcePathOnly, _ := cmd.Flags().GetBool(Flag_Compare_PrintSourcePathOnly)
	printZero, _ := cmd.Flags().GetBool(Flag_Compare_PrintZero)

	showExistsOnly, _ := cmd.Flags().GetBool(Flag_Compare_ShowExistsOnly)
	showMissingOnly, _ := cmd.Flags().GetBool(Flag_Compare_ShowMissingOnly)

	// check options
	if showExistsOnly && showMissingOnly {
		return 1, fmt.Errorf("Can't specify both -e and -m")
	}
	showMode := SHOW_ALWAYS
	if showExistsOnly {
		showMode = SHOW_EXISTS_ONLY
	} else if showMissingOnly {
		showMode = SHOW_MISSING_ONLY
	}

	srcHashData, err := core.LoadHashData(sourceFile)
	if err != nil {
		return 1, err
	}

	targetHashData, err := core.LoadHashData(targetFile)
	if err != nil {
		return 1, err
	}

	result, err := doCompare(srcHashData, targetHashData, printSourcePathOnly, printZero, showMode)
	return result, err
}

func doCompare(src *core.HashStore, target *core.HashStore, printSourcePathOnly bool, printZero bool, showMode int) (int, error) {
	sep := "\n"
	if printZero {
		sep = "\x00"
	}

	for _, hash := range src.Values() {
		sames := target.Get(hash.String())

		if len(sames) > 0 && showMode != SHOW_MISSING_ONLY {
			fmt.Print(makeResult(hash, sames, printSourcePathOnly))
			fmt.Print(sep)
		} else if len(sames) == 0 && showMode != SHOW_EXISTS_ONLY {
			fmt.Print(makeResult(hash, sames, printSourcePathOnly))
			fmt.Print(sep)
		}
	}
	return 0, nil
}

func makeResult(hash *core.Hash, sames []*core.Hash, printSourcePathOnly bool) string {
	if printSourcePathOnly {
		return hash.Path
	} else {
		result := fmt.Sprintf("%s\t%d", hash.Path, len(sames))
		for _, s := range sames {
			result += "\t" + s.Path
		}
		return result
	}
}
