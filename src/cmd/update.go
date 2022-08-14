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

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long:  ``,
	Run:   updateHash,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func updateHash(cmd *cobra.Command, args []string) {
	for _, v := range args {
		if err := doUpdateHash(v); err != nil {
			// TODO
			fmt.Printf("%s\n", err.Error())
		}
	}
}

func doUpdateHash(path string) error {
	hash, err := calcFileHash(path, crypto.SHA1)
	if err != nil {
		return err
	}

	fmt.Println(hash)

	return nil
}

func calcFileHash(path string, alg crypto.Hash) (string, error) {
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
