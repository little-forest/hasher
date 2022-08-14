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
	"crypto"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const Flag_Update_ForceUpdate = "force-update"

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long:  ``,
	Run:   updateHash,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolP(Flag_Update_ForceUpdate, "f", false, "Force update")
}

func updateHash(cmd *cobra.Command, args []string) {
	forceUpdate, _ := cmd.Flags().GetBool(Flag_Update_ForceUpdate)
	verbose, _ := cmd.Flags().GetBool(Flag_root_Verbose)

	alg := DefaultHashAlgorithm
	attrName := getPrefix(alg)

	for _, p := range args {
		changed, hash, err := doUpdateHash(p, alg, attrName, forceUpdate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			continue
		}
		if verbose {
			mark := ""
			if changed {
				mark = "*"
			}
			fmt.Fprintf(os.Stdout, "%s  %s %s\n", p, hash, mark)
		}
	}
}

func doUpdateHash(path string, alg crypto.Hash, attrName string, forceUpdate bool) (bool, string, error) {
	file, err := openFile(path)
	if err != nil {
		return false, "", err
	}
	defer file.Close()

	info, err := file.Stat()
	size := fmt.Sprint(info.Size())
	modTime := info.ModTime().Format(time.RFC3339Nano)

	var changed bool
	curHash := getXattr(file, attrName)
	if curHash != "" {
		if curSize := getXattr(file, Xattr_size); size != curSize {
			changed = true
		} else if curMtime := getXattr(file, Xattr_modifiedTime); modTime != curMtime {
			changed = true
		}
		if !forceUpdate && !changed {
			return false, curHash, nil
		}
	}

	hash, err := calcHashString(file, alg)
	if err != nil {
		return false, "", err
	}

	setXattr(file, attrName, hash)
	setXattr(file, Xattr_size, size)
	setXattr(file, Xattr_modifiedTime, modTime)

	return true, hash, nil
}

func getPrefix(alg crypto.Hash) string {
	return Xattr_prefix + "." + strings.ToLower(strings.Replace(alg.String(), "-", "", -1))
}
