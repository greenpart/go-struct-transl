package transl

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"testing"
)

func TestGetStructMeta(t *testing.T) {
	type goodStruct struct {
		Name         string `tr:"."`
		Kind         int
		Element      string `tr:"element"`
		Translations KeyLangValueMap
	}

	s := goodStruct{}
	meta, err := metas.getStructMeta(&s)

	assert.Nil(t, err)

	assert.Equal(t, &structMeta{
		translatable: false,
		getterIdx:    3,

		fields: []fieldMeta{
			fieldMeta{
				name:  "Name",
				key:   "Name",
				index: 0,
			},
			fieldMeta{
				name:  "Element",
				key:   "element",
				index: 2,
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

func TestGetStructMetaRegular(t *testing.T) {
	type regularStruct struct {
		Name string
		Kind int
	}

	s := regularStruct{}
	_, err := metas.getStructMeta(&s)

	if assert.NotNil(t, err) {
		assert.Equal(t, errors.New("Translate of struct without suitable fields"), err)
	}
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
	assert.Equal(t, &structMeta{
		translatable: true,
		getterIdx:    -1,
	}, meta)
}
