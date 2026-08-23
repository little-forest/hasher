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
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/little-forest/hasher/hashcore"
	"github.com/little-forest/hasher/internal/term"
)

func CleanPath(path string) (string, error) {
	if len(path) > 1 && path[0:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		path = homeDir + path[1:]
	}
	path = os.ExpandEnv(path)
	return filepath.Clean(path), nil
}

// CountAllFiles counts files under the specified file or directory.
func CountAllFiles(paths []string, verbose bool) int {
	threshold := 1000

	if verbose {
		fmt.Print(term.C_cyan.Apply("Counting files... "))
		term.HideCursor()
	}

	count := 0
	for _, p := range paths {
		t, err := hashcore.CheckFileType(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] %s\n", term.C_red.Apply("ERROR"), err.Error())
			continue
		}

		switch t {
		case hashcore.RegularFile:
			count++
			if verbose && (count%threshold) == 0 {
				fmt.Printf("\x1b7%d\x1b8", count)
			}

		case hashcore.Directory:
			err := WalkDir(p, func(f *os.File) error {
				count++
				if verbose && (count%threshold) == 0 {
					fmt.Printf("\x1b7%d\x1b8", count)
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "[%s] %s\n", term.C_red.Apply("ERROR"), err.Error())
				continue
			}
		default:
			fmt.Fprintf(os.Stderr, "[%s] Ignored : %s\n", term.C_yellow.Apply("WARN"), p)
		}

	}

	if verbose {
		fmt.Printf("%d\n", count)
		term.ShowCursor()
	}
	return count
}
