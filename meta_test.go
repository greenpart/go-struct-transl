package transl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

type goodStruct struct {
	Name         string `tr:"."`
	Kind         int
	Element      string `tr:"element"`
	Translations KeyLangValueMap
}

func TestGetStructMeta(t *testing.T) {
	s := goodStruct{}
	meta, err := metas.getStructMeta(&s)

	assert.Nil(t, err)

	assert.Equal(t, &StructMeta{
		Translatable: false,
		GetterIdx:    3,

		Fields: []fieldMeta{
			fieldMeta{
				Name:  "Name",
				Key:   "Name",
				Index: 0,
			},
			fieldMeta{
				Name:  "Element",
				Key:   "element",
				Index: 2,
			},
		},
	}, meta)
}

func TestGetStructMetaNoTranslations(t *testing.T) {
	type noTrStruct struct {
		Name string `tr:"."`
	}

	s := noTrStruct{}
	_, err := metas.getStructMeta(&s)

	if assert.NotNil(t, err) {
		assert.Equal(t, errors.New("Translate of struct without suitable fields"), err)
	}
}

type regularStruct struct {
	Name string
	Kind int
}

func TestGetStructMetaRegular(t *testing.T) {
	s := regularStruct{}
	_, err := metas.getStructMeta(&s)

	if assert.NotNil(t, err) {
		assert.Equal(t, errors.New("Translate of struct without suitable fields"), err)
	}
}

func TestGetStructMetaNotPointer(t *testing.T) {
	s := regularStruct{}
	_, err := metas.getStructMeta(s)

	if assert.NotNil(t, err) {
		assert.Equal(t, errors.New("Translate of non-pointer type"), err)
	}
}

func TestGetStructMetaImported(t *testing.T) {
	s := goodStruct{}
	v1, e1 := GetStructMeta(s)
	v2, e2 := metas.getStructMeta(s)
	assert.Equal(t, v1, v2)
	assert.Equal(t, e1, e2)
}

type translatableStruct struct {
	Name string
	Kind int
}

func (t *translatableStruct) Translate(preferred []language.Tag) error {
	return nil
}

func TestGetStructMetaTranslatable(t *testing.T) {
	s := translatableStruct{}
	meta, err := metas.getStructMeta(&s)

	assert.Nil(t, err)
	assert.Equal(t, &StructMeta{
		Translatable: true,
		GetterIdx:    -1,
	}, meta)
}
