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
	"strings"
)

type HashAlg struct {
	Alg      crypto.Hash
	AlgName  string
	AttrName string
}

func NewDefaultHashAlg(attrPrefix string) *HashAlg {
	return NewHashAlg(attrPrefix, crypto.SHA1)
}

func NewHashAlg(attrPrefix string, alg crypto.Hash) *HashAlg {
	algName := strings.ToLower(strings.ReplaceAll(alg.String(), "-", ""))
	hashAlg := &HashAlg{
		Alg:      alg,
		AlgName:  algName,
		AttrName: attrPrefix + "." + algName,
	}
	return hashAlg
}
