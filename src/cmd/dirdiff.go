package cmd

import (
	"fmt"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

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
}

func runDirDiff(cmd *cobra.Command, args []string) (int, error) {
	if len(args) < 2 {
		return 1, fmt.Errorf("too few argumrnts")
	}

	path1 := args[0]
	path2 := args[1]

	if err := checkDirectory(path1); err != nil {
		return 1, err
	}
	if err := checkDirectory(path2); err != nil {
		return 1, err
	}

	status, err := dirDiff(path1, path2, true)

	return status, err
}

func dirDiff(basePath string, targetPath string, verbose bool) (int, error) {
	// diff
	dirPairs, err := core.DirDiffRecursively(basePath, targetPath)

	// display
	for _, pair := range dirPairs {
		if pair.Status == core.BASE_ONLY {
			fmt.Println(C_cyan.Apply(fmt.Sprintf("[+] %s", pair.Path())))
			displayDir(pair.Base)
		} else if pair.Status == core.TARGET_ONLY {
			fmt.Println(C_pink.Apply(fmt.Sprintf("[-] %s", pair.Path())))
			displayDir(pair.Target)
		} else {
			// same
			if pair.Base.IsAllSame() {
				fmt.Println(C_gray.Apply(fmt.Sprintf("[=] %s", pair.Path())))
			} else {
				fmt.Printf("    %s\n", pair.Path())
			}
			displayDir(pair.Base)
		}
	}
	return 0, err
}

func displayDir(d *core.DirDiff) {
	for _, f := range d.GetSortedChildren() {
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