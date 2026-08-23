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

	"github.com/little-forest/hasher/hashcore"
	"github.com/little-forest/hasher/internal/fsutil"
	"github.com/little-forest/hasher/internal/hasher"
	"github.com/little-forest/hasher/internal/term"
	"github.com/spf13/cobra"
)

const Flag_find_NoHash = "no-hash"
const Flag_find_HasHash = "has-hash"
const Flag_find_File = "file"

var (
	findNoHash  bool
	findHasHash bool
	findFile    string
)

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
	RunE:         runFind,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().BoolVarP(&findNoHash, Flag_find_NoHash, "n", false, "Find files that has no hash value on XAttr")
	findCmd.Flags().BoolVarP(&findHasHash, Flag_find_HasHash, "e", false, "Find files that have hash value on XAttr")
	findCmd.Flags().StringVarP(&findFile, Flag_find_File, "f", "", "Find files that have same hash value as given file")
	findCmd.MarkFlagsMutuallyExclusive(Flag_find_NoHash, Flag_find_HasHash, Flag_find_File)
}

func runFind(cmd *cobra.Command, args []string) error {
	alg := hashcore.NewDefaultHashAlg()
	if findNoHash {
		w := &findNoHashWalker{Alg: alg}
		return fsutil.WalkDirsWithWalker(args, w)
	} else if findHasHash {
		w := &findHasHashWalker{Alg: alg}
		return fsutil.WalkDirsWithWalker(args, w)
	} else if findFile != "" {
		return findSameHashFile(alg, findFile, args)
	}
	return fmt.Errorf("invalid argument")
}

type findNoHashWalker struct {
	Alg *hashcore.HashAlg
}

func (w findNoHashWalker) Deal(f *os.File) error {
	hash, err := hashcore.GetHash(f.Name(), w.Alg)
	if err != nil {
		return err
	}
	if hash == nil {
		fmt.Printf("%s\n", f.Name())
	}
	return nil
}

type findHasHashWalker struct {
	Alg *hashcore.HashAlg
}

func (w findHasHashWalker) Deal(f *os.File) error {
	hash, err := hashcore.GetHash(f.Name(), w.Alg)
	if err != nil {
		return err
	}
	if hash != nil {
		fmt.Println(hash.Tsv())
	}
	return nil
}

type findSameHashWalker struct {
	Alg    *hashcore.HashAlg
	Source *hashcore.Hash
}

func (w findSameHashWalker) Deal(f *os.File) error {
	_, hash, err := hasher.UpdateHash(f.Name(), w.Alg, false)
	if err != nil {
		term.ShowWarn("failed to update hash : %s", err.Error())
	}
	if hash != nil && w.Source.HasSameHashValue(hash) {
		fmt.Println(hash.Tsv())
	}
	return nil
}

func findSameHashFile(alg *hashcore.HashAlg, srcPath string, targetDirs []string) error {
	if err := hashcore.EnsureRegularFile(srcPath); err != nil {
		return err
	}
	_, srcHash, err := hasher.UpdateHash(srcPath, alg, false)
	if err != nil {
		return err
	}

	w := &findSameHashWalker{Alg: alg, Source: srcHash}
	return fsutil.WalkDirsWithWalker(targetDirs, w)
}
