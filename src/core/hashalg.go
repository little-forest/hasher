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
	"crypto"
	_ "crypto/sha1"
	"strings"
)

type HashAlg struct {
	Alg      crypto.Hash
	AlgName  string
	AttrName string
}

func NewDefaultHashAlg() *HashAlg {
	return NewHashAlg(crypto.SHA1)
}

func NewHashAlg(alg crypto.Hash) *HashAlg {
	algName := strings.ToLower(strings.ReplaceAll(alg.String(), "-", ""))
	hashAlg := &HashAlg{
		Alg:      alg,
		AlgName:  algName,
		AttrName: Xattr_prefix + "." + algName,
	}
	return hashAlg
}

func NewHashAlgFromString(algName string) *HashAlg {
	switch algName {
	case "sha1":
		return NewHashAlg(crypto.SHA1)
	case "sha256":
		return NewHashAlg(crypto.SHA256)
	case "sha512":
		return NewHashAlg(crypto.SHA512)
	}
	return nil
}
