package cmd

import (
	"fmt"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

var MARK_SRC_EXISTS = C_yellow.Apply("[+]")
var MARK_SRC_NEWER = C_lime.Apply("[>]")
var MARK_SRC_OLDER = C_orange.Apply("[<]")
var MARK_NOT_SAME = C_pink.Apply("[~]")
var MARK_SRC_NOT_EXISTS = C_red.Apply("[-]")

// dirdiffCmd represents the dirdiff command
var dirdiffCmd = &cobra.Command{
	Use:   "dirdiff",
	Short: "A brief description of your command",
	Long:  ``,
	RunE:  statusWrapper.RunE(runDirDiff),
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

	// TODO
	// status, err := dirDiff(path1, path2, true)
	status, err := dirDiff_new(path1, path2, true)

	return status, err
}

func dirDiff_new(basePath string, targetPath string, verbose bool) (int, error) {
	dirPairs, err := core.DirDiffRecursive(basePath, targetPath)

	// display
	for _, pair := range dirPairs {
		if pair.Status == core.BASE_ONLY {
			fmt.Println(C_cyan.Apply(fmt.Sprintf("[+] %s", pair.Path())))
		} else if pair.Status == core.TARGET_ONLY {
			fmt.Println(C_pink.Apply(fmt.Sprintf("[-] %s", pair.Path())))
		} else {
			fmt.Printf("    %s\n", pair.Path())
		}

		if pair.Status == core.PAIR {
			for _, f := range pair.Base.GetSortedChildren() {
				col := getColorByStatus(f.Status)
				fmt.Println(col.Apply(fmt.Sprintf("      %s %s", f.StatusMark(), f.Basename)))
			}
		}
	}
	return 0, err
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

// func dirDiff(basePath string, targetPath string, verbose bool) (int, error) {
// 	basePath = filepath.Clean(basePath)
// 	targetPath = filepath.Clean(targetPath)
//
// 	alg := core.NewDefaultHashAlg()
//
// 	totalCount, err := CountFiles(basePath, verbose)
// 	if err != nil {
// 		return 1, err
// 	}
// 	count := 1
//
// 	targetList, _ := getTargetList(targetPath)
//
// 	if verbose {
// 		HideCursor()
// 	}
// 	err = filepath.WalkDir(basePath, func(path string, info fs.DirEntry, err error) error {
// 		if err != nil {
// 			return errors.Wrap(err, "failed to filepath.Walk")
// 		}
//
// 		relPath, _ := filepath.Rel(basePath, path)
//
// 		checkTargetPath := filepath.Join(targetPath, relPath)
// 		exists, _ := checkExistence(path, checkTargetPath, alg)
//
// 		if exists {
// 			targetList.Remove(checkTargetPath)
// 		}
//
// 		count++
//
// 		if !exists && info.IsDir() {
// 			// skip sub directories
// 			return filepath.SkipDir
// 		} else {
// 			return nil
// 		}
// 	})
// 	// if verbose {
// 	// 	fmt.Printf("\n")
// 	// 	showCursor()
// 	// }
//
// 	// print files on target side only
// 	for e := range targetList.Iterator().C {
// 		fmt.Println(MARK_SRC_NOT_EXISTS + " " + e)
// 	}
//
// 	fmt.Printf("%d / %d\n", count, totalCount)
//
// 	return 0, err
// }

// // return true if target file/directory exists
// func checkExistence(srcPath string, targetPath string, alg *core.HashAlg) (bool, error) {
// 	info, err := os.Stat(targetPath)
// 	exists := (err == nil)
//
// 	if !exists {
// 		// target file/directory not found
// 		fmt.Println(MARK_SRC_EXISTS + " " + srcPath)
// 		return false, nil
// 	}
//
// 	if info.IsDir() {
// 		// directory exists
// 		fmt.Println("    " + C_cyan.Apply(srcPath))
// 		return true, nil
// 	}
//
// 	isSame, err := compareHash(alg, srcPath, targetPath)
// 	if err != nil {
// 		// failed to check hash value
// 		fmt.Println(C_red.Apply("[?] " + srcPath + " " + err.Error()))
// 		return true, err
// 	} else {
// 		if isSame {
// 			// same file
// 			fmt.Println("    " + srcPath)
// 			return true, nil
// 		} else {
// 			// not same, judge which is newer
// 			statSrc, _ := os.Stat(srcPath)
// 			statTarget, _ := os.Stat(targetPath)
// 			srcModTime := statSrc.ModTime()
// 			targetModTime := statTarget.ModTime()
//
// 			if srcModTime.After(targetModTime) {
// 				fmt.Println(MARK_SRC_NEWER + " " + srcPath)
// 			} else if srcModTime.Before(targetModTime) {
// 				fmt.Println(MARK_SRC_OLDER + " " + srcPath)
// 			} else {
// 				fmt.Println(MARK_NOT_SAME + " " + srcPath)
// 			}
// 			return true, nil
// 		}
// 	}
// }

// // Compare hash value of givven files.
// // Two files are assumed to exist.
// // return true if two files are same.
// func compareHash(alg *core.HashAlg, path1 string, path2 string) (bool, error) {
// 	_, hash1, err := core.UpdateHash(path1, alg, false)
// 	if err != nil {
// 		return false, err
// 	}
//
// 	_, hash2, err := core.UpdateHash(path2, alg, false)
// 	if err != nil {
// 		return false, err
// 	}
//
// 	return (hash1 == hash2), nil
// }
//
// func getTargetList(path string) (mapset.Set[string], error) {
// 	targetFiles := mapset.NewSet[string]()
// 	err := filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
// 		if err != nil {
// 			return errors.Wrap(err, "failed to filepath.Walk")
// 		}
// 		targetFiles.Add(path)
// 		return nil
// 	})
//
// 	return targetFiles, err
// }

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
