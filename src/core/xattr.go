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
package core

import (
	"os"
	"strings"

	"github.com/pkg/xattr"
)

func GetXattr(file *os.File, attrName string) string {
	v, err := xattr.FGet(file, attrName)
	if err != nil {
		return ""
	}
	return string(v)
}

func SetXattr(file *os.File, attrName string, value string) error {
	return xattr.FSet(file, attrName, []byte(value))
}

func RemoveXattr(file *os.File, attrName string) error {
	return xattr.FRemove(file, attrName)
}

func ClearXattr(file *os.File) error {
	attrNames, err := xattr.FList(file)
	if err != nil {
		return err
	}

	for _, attrName := range attrNames {
		if strings.HasPrefix(attrName, Xattr_prefix) {
			if err := RemoveXattr(file, attrName); err != nil {
				return err
			}
		}
	}
	return nil
}
