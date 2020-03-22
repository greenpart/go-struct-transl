/*
Package transl translates struct fields and store translations
in the same struct.
*/
package transl

import (
	"database/sql/driver"
	"encoding/json"

	"golang.org/x/text/language"
)

// Translator is the interface that wraps the Translate self method.
//
// Translate changes value of TARGET to fit preferred languages.
// It returns any error encountered.
type Translator interface {
	Translate(target interface{}, preferred []language.Tag) error
}

// Translatable is the interface that wraps the Translate method.
//
// Translate changes value of CALLER to fit preferred languages.
// It returns any error encountered.
type Translatable interface {
	Translate(preferred []language.Tag) error
}

// TranslationsGetter is the interface that wraps the GetTranslations method.
//
// GetTranslations returns LangKeyValueMap with translations data.
//
// First field in struct which implements this interface will be used
// as a translations source
type TranslationsGetter interface {
	GetTranslations() KeyLangValueMap
}

// KeyLangValueMap holds map with translations
// Keys are fieldName first and language after.
// for example
// KeyLangValueMap{
//     "name": map[string]string{
//         "en": "Name",
//         "ru": "Имя",
//      },
//      "element": map[string]string{
//         "en": "water",
//         "ru": "вода",
//      },
// }
type KeyLangValueMap map[string]map[string]string

// GetTranslations implements TranslationsGetter interface
func (m KeyLangValueMap) GetTranslations() KeyLangValueMap {
	return m
}

// Scan unmarshals translations from JSON
func (m *KeyLangValueMap) Scan(value interface{}) error {
	if value == nil {
		m = &KeyLangValueMap{}
		return nil
	}
	return json.Unmarshal(value.([]byte), m)
}

// Value marshals translations to JSON
func (m KeyLangValueMap) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
