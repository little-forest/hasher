package hashcore

import (
	"crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The extended attribute names are part of the external interface and must not
// follow renames of this package. See SPECS.md SPEC-XATTR-001..006.
func TestXattrNamesAreFixed(t *testing.T) {
	assert.Equal(t, "user.hasher", Xattr_prefix)
	assert.Equal(t, "user.hasher.size", Xattr_size)
	assert.Equal(t, "user.hasher.mtime", Xattr_modifiedTime)
	assert.Equal(t, "user.hasher.htime", Xattr_hashCheckedTime)
}

func TestHashAlgAttrNamesAreFixed(t *testing.T) {
	assert.Equal(t, "user.hasher.sha1", NewDefaultHashAlg().AttrName)
	assert.Equal(t, "sha1", NewDefaultHashAlg().AlgName)

	assert.Equal(t, "user.hasher.sha1", NewHashAlg(crypto.SHA1).AttrName)
	assert.Equal(t, "user.hasher.sha256", NewHashAlg(crypto.SHA256).AttrName)
	assert.Equal(t, "user.hasher.sha512", NewHashAlg(crypto.SHA512).AttrName)

	assert.Equal(t, "user.hasher.sha256", NewHashAlgFromString("sha256").AttrName)
	assert.Nil(t, NewHashAlgFromString("md5"))
}
