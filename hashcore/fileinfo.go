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
package hashcore

import (
	"fmt"
	"os"
)

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
