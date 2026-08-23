package hasher

import (
	"testing"

	"github.com/little-forest/hasher/hashcore"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHash(t *testing.T) {
	alg := hashcore.NewDefaultHashAlg()
	path, expectedHash := makeSingleDummyFile(t, &alg.Alg)

	changed, hash, err := UpdateHash(path, alg, false)

	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash.String())
	assert.True(t, changed)
	// f, err := os.Open(path)
	// assert.NoError(t, err)

	// // check if hash value is saved to xattr
	// attrHash := hashcore.GetXattr(f, alg.AttrName)
	// assert.Equal(t, expectedHashValue, attrHash)
}
