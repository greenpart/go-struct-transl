package transl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetStructMeta(t *testing.T) {
	type goodStruct struct {
		Name         string `tr:"."`
		Kind         int
		Element      string `tr:"element"`
		Translations StringTable
	}

	s := goodStruct{}
	meta := metas.getStructMeta(s)

	assert.Equal(t, true, meta.valid)
	assert.Equal(t, 3, meta.trIndex)
	assert.Equal(t, 2, len(meta.fields))

	f := meta.fields[0]
	assert.Equal(t, "Name", f.name)
	assert.Equal(t, "Name", f.key)
	assert.Equal(t, 0, f.index)

	f = meta.fields[1]
	assert.Equal(t, "Element", f.name)
	assert.Equal(t, "element", f.key)
	assert.Equal(t, 2, f.index)
}

func TestGetStructMetaNoTranslations(t *testing.T) {
	type noTrStruct struct {
		Name string `tr:"."`
	}

	s := noTrStruct{}
	meta := metas.getStructMeta(s)

	assert.Equal(t, false, meta.valid)
	assert.Equal(t, -1, meta.trIndex)
	assert.Equal(t, 0, len(meta.fields))
}

func TestGetStructMetaRegular(t *testing.T) {
	type regularStruct struct {
		Name string
		Kind int
	}

	s := regularStruct{}
	meta := metas.getStructMeta(s)

	assert.Equal(t, false, meta.valid)
	assert.Equal(t, -1, meta.trIndex)
	assert.Equal(t, 0, len(meta.fields))
}
