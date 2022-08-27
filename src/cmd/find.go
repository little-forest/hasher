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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find [path ...]",
	Short: "find files which has hash attribute.",
	Long:  ``,
	RunE:  statusWrapper.RunE(runFind),
}

func init() {
	rootCmd.AddCommand(findCmd)
}

func runFind(cmd *cobra.Command, args []string) (int, error) {

	for _, rootDir := range args {
		err := doFind(rootDir)
		if err != nil {
			return 1, errors.Wrap(err, fmt.Sprintf("failed to walk : %s", rootDir))
		}
	}
	return 0, nil
}

func doFind(rootDir string) error {
	hashAlg := NewDefaultHashAlg(Xattr_prefix)
	err := filepath.WalkDir(rootDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		if hash, _ := getHashXattr(path, hashAlg); hash != "" {
			fmt.Printf("%s\t%s\n", path, hash)
		}
		return nil
	})
	return err
}

func getHashXattr(path string, hashAlg *HashAlg) (string, error) {
	f, err := openFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open : %s (%s)", path, err.Error())
		return "", err
	}

	hash := getXattr(f, hashAlg.AttrName)
	return hash, nil
}
