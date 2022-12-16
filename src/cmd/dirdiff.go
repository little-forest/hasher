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

	"github.com/little-forest/hasher/common"
	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

const Flag_DirDiff_showOnlyDifferences = "show-only-differences"

// dirdiffCmd represents the dirdiff command
var dirdiffCmd = &cobra.Command{
	Use:   "dirdiff BASE_DIR TARGET_DIR",
	Args:  cobra.ExactArgs(2),
	Short: "Recursively compares two directories and displays the differences.",
	Long: `Recursively compares two directories and displays the differences.
Each files are compared using hash values.

  [=] : same file
  [+] : added file
  [-] : removed file
  [>] : different file (base is newer)
  [<] : different file (target is newer)
  [~] : different file (modtime is same)
  [R] : renamed file
`,
	RunE: statusWrapper.RunE(runDirDiff),
}

func init() {
	rootCmd.AddCommand(dirdiffCmd)

	dirdiffCmd.Flags().BoolP(Flag_DirDiff_showOnlyDifferences, "d", false, "Show only differences")
}

func runDirDiff(cmd *cobra.Command, args []string) (int, error) {
	path1 := args[0]
	path2 := args[1]

	if err := checkDirectory(path1); err != nil {
		return 1, err
	}
	if err := checkDirectory(path2); err != nil {
		return 1, err
	}

	showOnlyDiff, _ := cmd.Flags().GetBool(Flag_DirDiff_showOnlyDifferences)

	status, err := dirDiff(path1, path2, showOnlyDiff, true)

	return status, err
}

func dirDiff(basePath string, targetPath string, showOnlyDiff bool, verbose bool) (int, error) {
	// diff
	dirPairs, err := core.DirDiffRecursively(basePath, targetPath)
	if err != nil {
		common.ShowErrorMsg("dirdiff failed : %s", err.Error())
		return 1, nil
	}

	// display
	for _, pair := range dirPairs {
		if pair.Status == core.BASE_ONLY {
			fmt.Println(C_cyan.Apply(fmt.Sprintf("[+] %s", pair.Path())))
			displayDir(pair.Base, showOnlyDiff)
		} else if pair.Status == core.TARGET_ONLY {
			fmt.Println(C_pink.Apply(fmt.Sprintf("[-] %s", pair.Path())))
			displayDir(pair.Target, showOnlyDiff)
		} else {
			// same
			if pair.Base.IsAllSame() && !showOnlyDiff {
				fmt.Println(C_gray.Apply(fmt.Sprintf("[=] %s", pair.Path())))
			} else {
				fmt.Printf("    %s\n", pair.Path())
			}
			displayDir(pair.Base, showOnlyDiff)
		}
	}

	// RESULT
	return 0, err
}

func displayDir(d *core.DirDiff, showOnlyDiff bool) {
	for _, f := range d.GetSortedChildren() {
		if showOnlyDiff && f.Status == core.SAME {
			continue
		}
		col := getColorByStatus(f.Status)

		msg := col.Apply(fmt.Sprintf("      %s %s", f.StatusMark(), f.Basename))
		if f.Status == core.RENAMED {
			msg += "  " + C_blue.Apply("<-->") + "  " + col.Apply(f.PairFileName)
		}
		fmt.Println(msg)
	}
}

func checkDirectory(path string) error {
	isDir, err := IsDirectory(path)
	if err != nil {
		return err
	}
	if !isDir {
		return fmt.Errorf("not a directory : %s", path)
	}
	return nil
}

func getColorByStatus(s core.DiffStatus) aec.ANSI {
	switch s {
	case core.ADDED:
		return C_lime
	case core.SAME:
		return C_gray
	case core.NOT_SAME_NEW:
		return C_orange
	case core.NOT_SAME_OLD:
		return C_orange
	case core.NOT_SAME:
		return C_orange
	case core.RENAMED:
		return C_yellow
	case core.REMOVED:
		return C_pink
	}
	return C_default
}
