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
	"strings"

	"github.com/pkg/xattr"
	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear hash attributes",
	Long:  ``,
	Run:   clear,
}

func init() {
	rootCmd.AddCommand(clearCmd)

}

func clear(cmd *cobra.Command, args []string) {
	// verbose, _ := cmd.Flags().GetBool(Flag_root_Verbose)

	for _, p := range args {
		doClear(p)
	}
}

func doClear(path string) error {
	file, err := openFile(path)
	if err != nil {
		return err
	}

	attrNames, err := xattr.FList(file)
	if err != nil {
		return err
	}

	for _, attrName := range attrNames {
		if strings.HasPrefix(attrName, Xattr_prefix) {
			removeXattr(file, attrName)
		}
	}
	return nil
}
