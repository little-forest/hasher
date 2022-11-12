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

const Flag_Duplication_Source = "source"
const Flag_Duplication_Target = "target"
const Flag_Duplication_ShowExistsOnly = "exists-only"
const Flag_Duplication_ShowMissingOnly = "missing-only"
const Flag_Duplication_PrintSourcePathOnly = "print-source-path-only"
const Flag_Duplication_PrintZero = "print0"

const SHOW_ALWAYS = 0
const SHOW_EXISTS_ONLY = 1
const SHOW_MISSING_ONLY = 2

// checkDuplicationCmd represents the compare command
var checkDuplicationCmd = &cobra.Command{
	Use:   "duplicate -s (HASH_LIST_TSV|SOURCE_DIR) -t (HASH_LIST_TSV|TARGET_DIR)",
	Short: "Check duplicated files",
	Long:  ``,
	RunE:  statusWrapper.RunE(runCheckDuplicated),
}

func init() {
	rootCmd.AddCommand(checkDuplicationCmd)

	checkDuplicationCmd.Flags().StringP(Flag_Duplication_Source, "s", "", "source hash file")
	checkDuplicationCmd.Flags().StringP(Flag_Duplication_Target, "t", "", "target hash file")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_ShowExistsOnly, "e", false, "show exist files only")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_ShowMissingOnly, "m", false, "show missing files only")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_PrintSourcePathOnly, "f", false, "print only source file path")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_PrintZero, "0", false, "separate by null character")
}

func runCheckDuplicated(cmd *cobra.Command, args []string) (int, error) {
	source, _ := cmd.Flags().GetString(Flag_Duplication_Source)
	target, _ := cmd.Flags().GetString(Flag_Duplication_Target)

	printSourcePathOnly, _ := cmd.Flags().GetBool(Flag_Duplication_PrintSourcePathOnly)
	printZero, _ := cmd.Flags().GetBool(Flag_Duplication_PrintZero)

	showExistsOnly, _ := cmd.Flags().GetBool(Flag_Duplication_ShowExistsOnly)
	showMissingOnly, _ := cmd.Flags().GetBool(Flag_Duplication_ShowMissingOnly)

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
	opt := CheckDuplicationOption{
		PrintSourcePathOnly: printSourcePathOnly,
		PrintZero:           printZero,
		ShowMode:            showMode,
	}

	alg := core.NewDefaultHashAlg()

	// make source hash store
	srcHashData, err := loadHashData(source, alg)
	if err != nil {
		return 1, err
	}

	// make target hash store
	targetHashData, err := loadHashData(target, alg)
	if err != nil {
		return 1, err
	}

	result, err := doCheckDuplication(srcHashData, targetHashData, opt)
	return result, err
}

func loadHashData(srcPath string, alg *core.HashAlg) (*core.HashStore, error) {
	var store *core.HashStore
	isDir, err := IsDirectory(srcPath)
	if err != nil {
		return nil, err
	}
	if isDir {
		store, err = core.MakeHashDataFromDirectory(srcPath, alg, false)
		if err != nil {
			return nil, err
		}
	} else {
		store, err = core.LoadHashData(srcPath)
		if err != nil {
			return nil, err
		}
	}
	return store, err
}

type CheckDuplicationOption struct {
	PrintSourcePathOnly bool
	PrintZero           bool
	ShowMode            int
}

func doCheckDuplication(src *core.HashStore, target *core.HashStore, opt CheckDuplicationOption) (int, error) {
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
