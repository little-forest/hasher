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

const Flag_Calc_NoShowPath = "no-show-path"

// calcCmd represents the calc command
var calcCmd = &cobra.Command{
	Use:   "calc",
	Short: "calc hash value and show",
	Long:  ``,
	Run:   calcHash,
}

func init() {
	rootCmd.AddCommand(calcCmd)

	calcCmd.Flags().BoolP(Flag_Calc_NoShowPath, "n", false, "don't show path")
}

func calcHash(cmd *cobra.Command, args []string) {
	for _, v := range args {
		hash, err := calcFileHash(v, DefaultHashAlgorithm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			continue
		}
		if f, _ := cmd.Flags().GetBool(Flag_Calc_NoShowPath); !f {
			fmt.Fprintf(os.Stdout, "%s  %s\n", hash, v)
		} else {
			fmt.Fprintf(os.Stdout, "%s\n", hash)
		}
	}
}

func openFile(path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("can't stat file : %s", path)
	}

	if info.Mode().IsDir() {
		return nil, fmt.Errorf("directory : %s", path)
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("non regular file : %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file. : %s", err.Error())
	}
	return file, nil
}

func calcFileHash(path string, alg crypto.Hash) (string, error) {
	r, err := openFile(path)
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
