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
	status := 0
	var errResult error
	for _, v := range args {
		err := showAttributes(v, NewDefaultHashAlg(Xattr_prefix))
		if err != nil {
			status = 1
			errResult = err
		}
	}
	return status, errResult
}

func showAttributes(path string, hashAlg *HashAlg) error {
	f, err := openFile(path)
	if err != nil {
		return err
	}

	hash := getXattr(f, hashAlg.AttrName)
	size := getXattr(f, Xattr_size)
	mTime := getXattr(f, Xattr_modifiedTime)
	fmt.Printf("%s\t%s\t%s\t%s\n", path, hash, size, mTime)

	return nil
}
