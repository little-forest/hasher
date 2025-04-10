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
	"strconv"
	"time"

	. "github.com/little-forest/hasher/common" // nolint:staticcheck
	"github.com/little-forest/hasher/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show extra attribute added by hasher",
	Long:  ``,
	RunE:  statusWrapper.RunE(runShow),
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) (int, error) {
	recuesive, _ := cmd.Flags().GetBool(Flag_root_Recursive)

	alg := core.NewDefaultHashAlg()

	showHeader()

	status := 0
	var errResult error
	for _, p := range args {
		isDir, err := IsDirectory(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			continue
		}

		if isDir {
			if !recuesive {
				// skip dir
				fmt.Fprintf(os.Stderr, "Skip directory : %s\n", p)
				continue
			}
			err = showAttributesRecursively(p, alg)
		} else {
			err = showAttributes(p, alg)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			status = 1
			errResult = err
		}
	}
	return status, errResult
}

func showAttributes(path string, hashAlg *core.HashAlg) error {
	f, err := OpenFile(path)
	if err != nil {
		return err
	}

	hash := core.GetXattr(f, hashAlg.AttrName)
	size := core.GetXattr(f, core.Xattr_size)

	mTime := getUnixTimeNano(f, core.Xattr_modifiedTime)
	hTime := getUnixTimeNano(f, core.Xattr_hashCheckedTime)

	fmt.Printf("%s\t%s\t%s\t%s\t%s\n", path, hash, size, mTime, hTime)

	return nil
}

func showHeader() {
	fmt.Println("Path\tHashValue\tSize\tModTime\tCheckedTime")
}

func getUnixTimeNano(f *os.File, attrName string) string {
	timeStr := ""
	t, err := strconv.ParseInt(core.GetXattr(f, attrName), 10, 64)
	if err == nil {
		timeStr = time.Unix(0, t).Format(time.RFC3339Nano)
	}
	return timeStr
}

func showAttributesRecursively(dirPath string, hashAlg *core.HashAlg) error {
	err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		showErr := showAttributes(path, core.NewDefaultHashAlg())
		if showErr != nil {
			fmt.Fprintf(os.Stderr, "%s\n", showErr.Error())
		}

		return nil
	})
	return err
}
