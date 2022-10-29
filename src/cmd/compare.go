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

	. "github.com/little-forest/hasher/common"
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
	Use:   "compare -s (HASH_LIST_TSV|SOURCE_DIR) -t TARGET_DIR",
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
	source, _ := cmd.Flags().GetString(Flag_Compare_Source)
	target, _ := cmd.Flags().GetString(Flag_Compare_Target)

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
	opt := CompareOption{
		PrintSourcePathOnly: printSourcePathOnly,
		PrintZero:           printZero,
		ShowMode:            showMode,
	}

	// make source hash store
	var srcHashData *core.HashStore
	alg := core.NewDefaultHashAlg()
	isDir, err := IsDirectory(source)
	if err != nil {
		return 1, err
	}
	if isDir {
		srcHashData, err = core.MakeHashDataFromDirectory(source, alg, false)
		if err != nil {
			return 1, err
		}
	} else {
		srcHashData, err = core.LoadHashData(source)
		if err != nil {
			return 1, err
		}
	}

	// make target hash store
	targetHashData, err := core.LoadHashData(target)
	if err != nil {
		return 1, err
	}

	result, err := doCompare(srcHashData, targetHashData, opt)
	return result, err
}

type CompareOption struct {
	PrintSourcePathOnly bool
	PrintZero           bool
	ShowMode            int
}

func doCompare(src *core.HashStore, target *core.HashStore, opt CompareOption) (int, error) {
	sep := "\n"
	if opt.PrintZero {
		sep = "\x00"
	}

	for _, hash := range src.Values() {
		sames := target.Get(hash.String())

		if len(sames) > 0 && opt.ShowMode != SHOW_MISSING_ONLY {
			fmt.Print(makeResult(hash, sames, opt.PrintSourcePathOnly))
			fmt.Print(sep)
		} else if len(sames) == 0 && opt.ShowMode != SHOW_EXISTS_ONLY {
			fmt.Print(makeResult(hash, sames, opt.PrintSourcePathOnly))
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
