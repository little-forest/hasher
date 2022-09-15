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

	"github.com/morikuni/aec"
	"github.com/pkg/errors"
	"github.com/pkg/xattr"
)

// ANSI escape sequences
//
//	see: https://github.com/morikuni/aec
var C_green = aec.EmptyBuilder.GreenF().Bold().ANSI
var C_red = aec.EmptyBuilder.RedF().Bold().ANSI
var C_lred = aec.EmptyBuilder.LightRedF().ANSI
var C_blue = aec.EmptyBuilder.BlueF().Bold().ANSI
var C_yellow = aec.EmptyBuilder.YellowF().Bold().ANSI
var C_cyan = aec.EmptyBuilder.CyanF().Bold().ANSI
var C_white = aec.EmptyBuilder.WhiteF().Bold().ANSI

var C_pink = aec.EmptyBuilder.Color8BitF(218).ANSI
var C_malibu = aec.EmptyBuilder.Color8BitF(74).ANSI
var C_orange = aec.EmptyBuilder.Color8BitF(214).ANSI
var C_darkorange3 = aec.EmptyBuilder.Color8BitF(166).ANSI
var C_lime = aec.EmptyBuilder.Color8BitF(10).ANSI

func openFile(path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("can't stat file : %s", path)
	}

	if info.Mode().IsDir() {
		return nil, fmt.Errorf("directory can't open : %s", path)
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

func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)

	if err != nil {
		return false, fmt.Errorf("can't stat file : %s", path)
	}

	return info.Mode().IsDir(), nil
}

func getXattr(file *os.File, attrName string) string {
	v, err := xattr.FGet(file, attrName)
	if err != nil {
		return ""
	}
	return string(v)
}

func setXattr(file *os.File, attrName string, value string) error {
	return xattr.FSet(file, attrName, []byte(value))
}

func removeXattr(file *os.File, attrName string) error {
	return xattr.FRemove(file, attrName)
}

func cleanPath(path string) (string, error) {
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

func countFiles(path string, verbose bool) (int, error) {
	threshold := 1000

	if verbose {
		fmt.Printf("Counting files : %s ... ", path)
		hideCursor()
	}

	count := 0
	err := filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		if verbose && (count%threshold) == 0 {
			fmt.Printf("\x1b7%d\x1b8", count)
		}
		count++

		return nil
	})
	if verbose {
		fmt.Printf("%d\n", count)
		showCursor()
	}
	return count, err
}

func showCursor() {
	fmt.Print("\x1b[?25h")
}

func hideCursor() {
	fmt.Print("\x1b[?25l")
}
