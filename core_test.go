package transl_test

import (
	"github.com/greenpart/go-struct-transl"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyLangValueMapJsonScan(t *testing.T) {
	st := transl.KeyLangValueMap{}
	err := st.Scan([]uint8(`{}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, transl.KeyLangValueMap{}, st)

	st = transl.KeyLangValueMap{}
	err = st.Scan([]uint8(`{"name":{}}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, transl.KeyLangValueMap{"name": map[string]string{}}, st)

	st = transl.KeyLangValueMap{}
	err = st.Scan([]uint8(`{"name":{"en":"Bob"}}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, transl.KeyLangValueMap{"name": map[string]string{"en": "Bob"}}, st)

	st = transl.KeyLangValueMap{}
	err = st.Scan([]uint8(`{"name":{"en":"Bob"}, "element":{"ru": "earth"}}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, transl.KeyLangValueMap{"name": map[string]string{"en": "Bob"}, "element": map[string]string{"ru": "earth"}}, st)
}

func TestKeyLangValueMapJsonValue(t *testing.T) {
	st := transl.KeyLangValueMap{}
	v, err := st.Value()
	assert.Equal(t, nil, err)
	assert.Equal(t, "{}", v)

	st = transl.KeyLangValueMap{"name": map[string]string{}}
	v, err = st.Value()
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"name":{}}`, v)
}

func TestKeyLangValueMapGetTranslations(t *testing.T) {
	st := transl.KeyLangValueMap{}
	assert.Equal(t, st, st.GetTranslations())
}
