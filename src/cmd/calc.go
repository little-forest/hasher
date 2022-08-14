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
	"io"
	"os"

	"github.com/spf13/cobra"
)

const HashBufSize = 256 * 1024
const DefaultHashAlgorithm = crypto.SHA1

// calcCmd represents the calc command
var calcCmd = &cobra.Command{
	Use:   "calc",
	Short: "A brief description of your command",
	Long: ``,
	Run: calcHash,
}

func init() {
	rootCmd.AddCommand(calcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// calcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// calcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func calcHash(cmd *cobra.Command, args []string) {
	for _, v := range args {
		hash, err := calcFileHash(v, DefaultHashAlgorithm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			continue
		} 
		fmt.Fprintf(os.Stdout, "%s  %s\n", hash, v)
	}
}

func calcFileHash(path string, alg crypto.Hash) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("can't stat file : %s", path)
	}

	if info.Mode().IsDir() {
		return "", fmt.Errorf("directory : %s", path)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("non regular file : %s", path)
	}

	r, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file. : %s", err.Error())
	}
	defer r.Close()

	hash, err := calcHashString(r, alg)
	return hash, err
}

func calcHashString(r io.Reader, alg crypto.Hash) (string, error) {
	if !alg.Available() {
		return "", fmt.Errorf("no implementation")
	}

	hash := alg.New()
	if _, err := io.CopyBuffer(hash, r, make([]byte, HashBufSize)); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
