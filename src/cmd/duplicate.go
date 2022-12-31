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

const (
	SHOW_ALWAYS       = iota + 1
	SHOW_EXISTS_ONLY  = 1
	SHOW_MISSING_ONLY = 2
)

type checkDuplicationOption struct {
	HashAlg             *core.HashAlg
	PrintSourcePathOnly bool
	PrintZero           bool
	ShowMode            int
	Source              []string
	Target              []string
}

// checkDuplicationCmd represents the compare command
var checkDuplicationCmd = &cobra.Command{
	Use:   "duplicate -s (HASH_LIST_TSV|SOURCE_DIR) -t (HASH_LIST_TSV|TARGET_DIR)",
	Short: "Check duplicated files",
	Example: `
  (1) Find each file in SOURCE_DIR exists in TARGET_DIRs...
        hasher duplicate -s SOURCE_DIR TARGET_DIR...

  (2) Find each file in SOURCE_DIRs exists in TARGET_DIR
        hasher duplicate -t TARGET_DIR SOURCE_DIRs

  Instead of directories, you can also specify a TSV file output by the list-hash sub-command.
  Cannot use -s and -t options at the same time.
`,
	RunE: statusWrapper.RunE(runCheckDuplicated),
	Args: func(cmd *cobra.Command, args []string) error {
		showExistsOnly, _ := cmd.Flags().GetBool(Flag_Duplication_ShowExistsOnly)
		showMissingOnly, _ := cmd.Flags().GetBool(Flag_Duplication_ShowMissingOnly)

		if showExistsOnly && showMissingOnly {
			return fmt.Errorf("can't specify both -e and -m option")
		}

		source, _ := cmd.Flags().GetString(Flag_Duplication_Source)
		target, _ := cmd.Flags().GetString(Flag_Duplication_Target)
		if source != "" && target != "" {
			return fmt.Errorf("can't specify both -s and -t option")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkDuplicationCmd)

	checkDuplicationCmd.Flags().StringP(Flag_Duplication_Source, "s", "", "source hash file or directory")
	checkDuplicationCmd.Flags().StringP(Flag_Duplication_Target, "t", "", "target hash file or directory")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_ShowExistsOnly, "e", false, "show exist files only")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_ShowMissingOnly, "m", false, "show missing files only")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_PrintSourcePathOnly, "f", false, "print only source file path")
	checkDuplicationCmd.Flags().BoolP(Flag_Duplication_PrintZero, "0", false, "separate by null character")
}

func newCkeckDuplicationOption(cmd *cobra.Command, args []string) checkDuplicationOption {
	showExistsOnly, _ := cmd.Flags().GetBool(Flag_Duplication_ShowExistsOnly)
	showMissingOnly, _ := cmd.Flags().GetBool(Flag_Duplication_ShowMissingOnly)

	showMode := SHOW_ALWAYS
	if showExistsOnly {
		showMode = SHOW_EXISTS_ONLY
	} else if showMissingOnly {
		showMode = SHOW_MISSING_ONLY
	}

	printSourcePathOnly, _ := cmd.Flags().GetBool(Flag_Duplication_PrintSourcePathOnly)
	printZero, _ := cmd.Flags().GetBool(Flag_Duplication_PrintZero)

	opt := checkDuplicationOption{
		HashAlg:             core.NewDefaultHashAlg(),
		PrintSourcePathOnly: printSourcePathOnly,
		PrintZero:           printZero,
		ShowMode:            showMode,
	}

	// set target and source
	optSource, _ := cmd.Flags().GetString(Flag_Duplication_Source)
	optTarget, _ := cmd.Flags().GetString(Flag_Duplication_Target)

	if optSource != "" {
		// multiple target
		opt.Source = []string{optSource}
		opt.Target = make([]string, len(args))
		copy(opt.Target, args)
	} else {
		// multiple source
		opt.Target = []string{optTarget}
		opt.Source = make([]string, len(args))
		copy(opt.Source, args)
	}

	return opt
}

func runCheckDuplicated(cmd *cobra.Command, args []string) (int, error) {
	opt := newCkeckDuplicationOption(cmd, args)

	// make source hash store
	srcHashData, err := loadHashData(opt.Source, opt.HashAlg)
	if err != nil {
		return 1, err
	}

	// make target hash store
	targetHashData, err := loadHashData(opt.Target, opt.HashAlg)
	if err != nil {
		return 1, err
	}

	result, err := doCheckDuplication(srcHashData, targetHashData, opt)
	return result, err
}

func loadHashData(srcPaths []string, alg *core.HashAlg) (*core.HashStore, error) {
	store := core.NewHashStore()
	for _, p := range srcPaths {
		isDir, err := IsDirectory(p)
		if err != nil {
			return nil, err
		}

		if isDir {
			err = store.AppendHashDataFromDirectory(p, alg, false)
			if err != nil {
				return nil, err
			}
		} else {
			err = store.LoadHashData(p)
			if err != nil {
				return nil, err
			}
		}
	}
	return store, nil
}

func doCheckDuplication(src *core.HashStore, target *core.HashStore, opt checkDuplicationOption) (int, error) {
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
