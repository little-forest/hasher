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
package common

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/morikuni/aec"
	"github.com/pkg/errors"
)

// ANSI escape sequences
//
//	see: https://github.com/morikuni/aec
var C_default = aec.EmptyBuilder.DefaultF().ANSI
var C_green = aec.EmptyBuilder.GreenF().ANSI
var C_red = aec.EmptyBuilder.RedF().ANSI
var C_lred = aec.EmptyBuilder.LightRedF().ANSI
var C_blue = aec.EmptyBuilder.BlueF().ANSI
var C_yellow = aec.EmptyBuilder.YellowF().ANSI
var C_cyan = aec.EmptyBuilder.CyanF().ANSI
var C_white = aec.EmptyBuilder.WhiteF().ANSI
var C_gray = aec.EmptyBuilder.Color8BitF(8).ANSI

var C_pink = aec.EmptyBuilder.Color8BitF(218).ANSI
var C_malibu = aec.EmptyBuilder.Color8BitF(74).ANSI
var C_orange = aec.EmptyBuilder.Color8BitF(214).ANSI
var C_darkorange3 = aec.EmptyBuilder.Color8BitF(166).ANSI
var C_lime = aec.EmptyBuilder.Color8BitF(10).ANSI

var Mark_OK = fmt.Sprintf("[%s]", C_green.Apply("OK"))
var Mark_Failed = fmt.Sprintf("[%s]", C_red.Apply("FAILED"))

func ShowWarn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "["+C_yellow.Apply("WARNING")+"] "+format+"\n", args...)
}

func ShowError(err error) {
	fmt.Fprintln(os.Stderr, "["+C_red.Apply("ERROR")+"] "+err.Error())
}

func ShowErrorMsg(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "["+C_red.Apply("ERROR")+"] "+format+"\n", args...)
}

type FileType int

const (
	Unknown FileType = iota + 1
	RegularFile
	Directory
	SymbolicLink
)

func CheckFileType(path string) (FileType, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Unknown, err
	}
	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		return SymbolicLink, nil
	}
	if info.Mode().IsDir() {
		return Directory, nil
	}
	if info.Mode().IsRegular() {
		return RegularFile, nil
	}
	return Unknown, fmt.Errorf("non regular file : %s", path)
}

func OpenFile(path string) (*os.File, error) {
	ftype, err := CheckFileType(path)
	if err != nil {
		return nil, err
	}

	switch ftype {
	case RegularFile:
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open file. : %s", err.Error())
		}
		return file, nil
	case Directory:
		return nil, fmt.Errorf("directory can't open : %s", path)
	case SymbolicLink:
		return nil, fmt.Errorf("symbolic link can't open : %s", path)
	default:
		return nil, fmt.Errorf("non regular file : %s", path)
	}
}

func IsDirectory(path string) (bool, error) {
	ftype, err := CheckFileType(path)
	if err != nil {
		return false, err
	}

	return ftype == Directory, nil
}

func EnsureDirectory(path string) error {
	isDir, err := IsDirectory(path)
	if err != nil {
		return err
	}
	if !isDir {
		return fmt.Errorf("not a directory : %s", path)
	}
	return nil
}

func EnsureRegularFile(path string) error {
	ftype, err := CheckFileType(path)
	if err != nil {
		return err
	}
	if ftype == Directory {
		return fmt.Errorf("directory : %s", path)
	}
	if ftype == SymbolicLink {
		return fmt.Errorf("symboliclink : %s", path)
	}
	return nil
}

func IsSymbolicLink(path string) (bool, error) {
	ftype, err := CheckFileType(path)
	if err != nil {
		return false, err
	}

	return ftype == SymbolicLink, nil
}

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

func CountAllFiles(paths []string, verbose bool) (int, error) {
	total := 0
	for _, p := range paths {
		files, err := CountFiles(p, verbose)
		total += files
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func CountFiles(path string, verbose bool) (int, error) {
	threshold := 1000

	if verbose {
		fmt.Printf("Counting files : %s ... ", path)
		HideCursor()
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
		ShowCursor()
	}
	return count, err
}

func ShowCursor() {
	fmt.Print("\x1b[?25h")
}

func HideCursor() {
	fmt.Print("\x1b[?25l")
}
