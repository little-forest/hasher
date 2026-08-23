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

	"github.com/little-forest/hasher/hashcore"
	"github.com/little-forest/hasher/internal/hasher"
	"github.com/spf13/cobra"
)

const Flag_duplicate_Source = "source"
const Flag_duplicate_Target = "target"
const Flag_duplicate_ShowExistsOnly = "exists-only"
const Flag_duplicate_ShowMissingOnly = "missing-only"
const Flag_duplicate_PrintSourcePathOnly = "print-source-path-only"
const Flag_duplicate_PrintZero = "print0"

var (
	duplicateSource              string
	duplicateTarget              string
	duplicateShowExistsOnly      bool
	duplicateShowMissingOnly     bool
	duplicatePrintSourcePathOnly bool
	duplicatePrintZero           bool
)

const (
	SHOW_ALWAYS = iota + 1
	SHOW_EXISTS_ONLY
	SHOW_MISSING_ONLY
)

type checkDuplicationOption struct {
	HashAlg             *hashcore.HashAlg
	Source              []string
	Target              []string
	ShowMode            int
	PrintSourcePathOnly bool
	PrintZero           bool
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
	RunE:         runCheckDuplicated,
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if duplicateShowExistsOnly && duplicateShowMissingOnly {
			return fmt.Errorf("can't specify both -e and -m option")
		}

		if duplicateSource != "" && duplicateTarget != "" {
			return fmt.Errorf("can't specify both -s and -t option")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkDuplicationCmd)

	checkDuplicationCmd.Flags().StringVarP(&duplicateSource, Flag_duplicate_Source, "s", "", "source hash file or directory")
	checkDuplicationCmd.Flags().StringVarP(&duplicateTarget, Flag_duplicate_Target, "t", "", "target hash file or directory")
	checkDuplicationCmd.Flags().BoolVarP(&duplicateShowExistsOnly, Flag_duplicate_ShowExistsOnly, "e", false, "show exist files only")
	checkDuplicationCmd.Flags().BoolVarP(&duplicateShowMissingOnly, Flag_duplicate_ShowMissingOnly, "m", false, "show missing files only")
	checkDuplicationCmd.Flags().BoolVarP(&duplicatePrintSourcePathOnly, Flag_duplicate_PrintSourcePathOnly, "f", false, "print only source file path")
	checkDuplicationCmd.Flags().BoolVarP(&duplicatePrintZero, Flag_duplicate_PrintZero, "0", false, "separate by null character")
}

func newCkeckDuplicationOption(cmd *cobra.Command, args []string) checkDuplicationOption {
	showMode := SHOW_ALWAYS
	if duplicateShowExistsOnly {
		showMode = SHOW_EXISTS_ONLY
	} else if duplicateShowMissingOnly {
		showMode = SHOW_MISSING_ONLY
	}

	opt := checkDuplicationOption{
		HashAlg:             hashcore.NewDefaultHashAlg(),
		PrintSourcePathOnly: duplicatePrintSourcePathOnly,
		PrintZero:           duplicatePrintZero,
		ShowMode:            showMode,
	}

	// set target and source
	if duplicateSource != "" {
		// multiple target
		opt.Source = []string{duplicateSource}
		opt.Target = make([]string, len(args))
		copy(opt.Target, args)
	} else {
		// multiple source
		opt.Target = []string{duplicateTarget}
		opt.Source = make([]string, len(args))
		copy(opt.Source, args)
	}

	return opt
}

func runCheckDuplicated(cmd *cobra.Command, args []string) error {
	opt := newCkeckDuplicationOption(cmd, args)

	// make source hash store
	srcHashData, err := loadHashData(opt.Source, opt.HashAlg)
	if err != nil {
		return err
	}

	// make target hash store
	targetHashData, err := loadHashData(opt.Target, opt.HashAlg)
	if err != nil {
		return err
	}

	return doCheckDuplication(srcHashData, targetHashData, opt)
}

func loadHashData(srcPaths []string, alg *hashcore.HashAlg) (*hasher.HashStore, error) {
	store := hasher.NewHashStore()
	for _, p := range srcPaths {
		isDir, err := hashcore.IsDirectory(p)
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

func doCheckDuplication(src *hasher.HashStore, target *hasher.HashStore, opt checkDuplicationOption) error {
	sep := "\n"
	if opt.PrintZero {
		sep = "\x00"
	}

	for _, hash := range src.Values() {
		sames := target.Get(hash.String())
		hasSame := len(sames) > 0

		// Src Target | no-opt missing-only existing-only
		// -----------+-----------------------------------
		//  o    o    |   o         -            o
		//  o    -    |   o         o            -
		//  -    o    |  N/A       N/A          N/A
		if hasSame && opt.ShowMode != SHOW_MISSING_ONLY {
			fmt.Print(makeResult(hash, sames, opt.PrintSourcePathOnly))
			fmt.Print(sep)
		} else if !hasSame && opt.ShowMode != SHOW_EXISTS_ONLY {
			fmt.Print(makeResult(hash, sames, opt.PrintSourcePathOnly))
			fmt.Print(sep)
		}
	}
	return nil
}

func makeResult(hash *hashcore.Hash, sames []*hashcore.Hash, printSourcePathOnly bool) string {
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
