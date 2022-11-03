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

	"github.com/little-forest/hasher/core"
	"github.com/spf13/cobra"
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare FILE1 FILE2",
	Short: "Compare if two files are identical",
	Long:  ``,
	RunE:  statusWrapper.RunE(runCompare),
}

func init() {
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, args []string) (int, error) {
	if len(args) < 2 {
		return -1, fmt.Errorf("too few arguments")
	}

	result, err := compare(args[0], args[1])
	if err != nil {
		return 1, nil
	}

	if result {
		return 0, nil
	} else {
		return 1, nil
	}
}

/*
Return true if given two failes have same hash value.
*/
func compare(path1 string, path2 string) (bool, error) {
	hashAlg := core.NewDefaultHashAlg()
	_, hash1, err := core.UpdateHash(path1, hashAlg, false)
	if err != nil {
		return false, err
	}

	_, hash2, err := core.UpdateHash(path2, hashAlg, false)
	if err != nil {
		return false, err
	}

	return (hash1.String() == hash2.String()), nil
}
