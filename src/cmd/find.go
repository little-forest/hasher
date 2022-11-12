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
	"os"

	. "github.com/little-forest/hasher/common"
	"github.com/little-forest/hasher/core"
	"github.com/spf13/cobra"
)

const Flag_Find_NoHash = "no-hash"
const Flag_Find_HasHash = "has-hash"
const Flag_Find_File = "file"

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find [path ...]",
	Short: "Find files which has hash attribute.",
	Long:  ``,
	Example: `
  (1) Find files that have no hash value on XAttr from directories 
        hasher find -n DIR...

  (2) Find files that have hash value on XAttr from directories
        hasher find -e DIR...

  (3) Find files that have same hash value as given SRCFILE from directories
        hasher find -f SRCFILE DIR...
`,
	RunE: statusWrapper.RunE(runFind),
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().BoolP(Flag_Find_NoHash, "n", false, "Find files that has no hash value on XAttr")
	findCmd.Flags().BoolP(Flag_Find_HasHash, "e", false, "Find files that have hash value on XAttr")
	findCmd.Flags().StringP(Flag_Find_File, "f", "", "Find files that have same hash value as given file")
	findCmd.MarkFlagsMutuallyExclusive(Flag_Find_NoHash, Flag_Find_HasHash, Flag_Find_File)
}

func runFind(cmd *cobra.Command, args []string) (int, error) {
	findNoHash, _ := cmd.Flags().GetBool(Flag_Find_NoHash)
	findHasHash, _ := cmd.Flags().GetBool(Flag_Find_HasHash)
	srcFile, _ := cmd.Flags().GetString(Flag_Find_File)

	alg := core.NewDefaultHashAlg()
	if findNoHash {
		w := &findNoHashWalker{Alg: alg}
		if err := WalkDirsWithWalker(args, w); err != nil {
			return 1, err
		} else {
			return 0, nil
		}
	} else if findHasHash {
		w := &findHasHashWalker{Alg: alg}
		if err := WalkDirsWithWalker(args, w); err != nil {
			return 1, err
		} else {
			return 0, nil
		}
	} else if srcFile != "" {
		if err := findSameHashFile(alg, srcFile, args); err != nil {
			return 1, err
		} else {
			return 0, nil
		}
	}
	return 1, fmt.Errorf("Invalid argument")
}

type findNoHashWalker struct {
	Alg *core.HashAlg
}

func (w findNoHashWalker) Deal(f *os.File) error {
	hash, err := core.GetHash(f.Name(), w.Alg)
	if err != nil {
		return err
	}
	if hash == nil {
		fmt.Printf("%s\n", f.Name())
	}
	return nil
}

type findHasHashWalker struct {
	Alg *core.HashAlg
}

func (w findHasHashWalker) Deal(f *os.File) error {
	hash, err := core.GetHash(f.Name(), w.Alg)
	if err != nil {
		return err
	}
	if hash != nil {
		fmt.Println(hash.Tsv())
	}
	return nil
}

type findSameHashWalker struct {
	Alg    *core.HashAlg
	Source *core.Hash
}

func (w findSameHashWalker) Deal(f *os.File) error {
	_, hash, err := core.UpdateHash(f.Name(), w.Alg, false)
	if err != nil {
		return err
	}
	if hash != nil && w.Source.HasSameHashValue(hash) {
		fmt.Println(hash.Tsv())
	}
	return nil
}

func findSameHashFile(alg *core.HashAlg, srcPath string, targetDirs []string) error {
	if err := EnsureRegularFile(srcPath); err != nil {
		return err
	}
	_, srcHash, err := core.UpdateHash(srcPath, alg, false)
	if err != nil {
		return err
	}

	w := &findSameHashWalker{Alg: alg, Source: srcHash}
	return WalkDirsWithWalker(targetDirs, w)
}
