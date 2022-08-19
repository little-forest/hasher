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
	"os"

	"github.com/pkg/xattr"
)

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
