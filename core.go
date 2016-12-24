/*
Package transl translates struct fields and store translations
in the same struct.
*/
package transl

import (
	"errors"
	// "fmt"
	"database/sql/driver"
	"encoding/json"
	"golang.org/x/text/language"
	"reflect"
	"sync"
)

var defaultLanguageString = "en"
var defaultLanguageTag = language.English

// SetDefaults redefines default language string and tag
func SetDefaults(str string, tag language.Tag) {
	defaultLanguageString = str
	defaultLanguageTag = tag
}

// Translater is the interface that wraps the Translate self method.
//
// Translate changes value of target to fit preferred languages.
// It returns any error encountered.
type Translater interface {
	Translate(target interface{}, preferred []language.Tag) error
}

// Translatable is the interface that wraps the Translate method.
//
// Translate changes value of caller to fit preferred languages.
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

// Translations holds map with language translations data
// for example
// LangKeyValueMap{
//     "en": map[string]string{
//         "name": "John",
//         "beloved": "Yoko",
//     },
// }
// type LangKeyValueMap map[string]map[string]string

// // GetTranslations implements TranslationsGetter interface by returning its value
// func (m LangKeyValueMap) GetTranslations() *KeyLangValueMap {
// 	// TODO implement index order changes
// 	return &KeyLangValueMap{}
// }

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

func (m KeyLangValueMap) GetTranslations() KeyLangValueMap {
	return m
}

// Scan unmarshals translations from JSON
func (m *KeyLangValueMap) Scan(value interface{}) error {
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

// Translate applies translation to target.
//
// If target implements Translatable interface
// this function calls Translate method on target.
func Translate(target interface{}, preferred []language.Tag) error {
	meta, err := metas.getStructMeta(target)
	if err != nil {
		return err
	}

	// Translate target with its Translate method
	// if it implements Translatable interface
	if meta.translatable {
		tr := target.(Translatable)
		return tr.Translate(preferred)
	}

	if meta.getterIdx >= 0 {
		return translateStructWithGetterField(target, preferred, meta)
	}

	return errors.New("Translate of unsupported type")
}

func translateStructWithGetterField(target interface{}, preferred []language.Tag, meta *structMeta) error {
	structValue := reflect.Indirect(reflect.ValueOf(target))

	getter := structValue.Field(meta.getterIdx).Interface().(TranslationsGetter)

	translations := getter.GetTranslations()
	// Empty translations Don't produce error
	if len(translations) == 0 {
		return nil
	}

	if len(preferred) == 0 {
		preferred = []language.Tag{defaultLanguageTag}
	}

	for _, trF := range meta.fields {
		f := structValue.Field(trF.index)
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			translateField(f, trF.key, translations, preferred)
		}
	}

	return nil
}

func translateField(field reflect.Value, fieldKey string, translations KeyLangValueMap, preferred []language.Tag) {
	matcher := getMatcher(fieldKey, translations)
	effectiveLang, _, _ := matcher.Match(preferred...)
	field.SetString(translations[fieldKey][effectiveLang.String()])
}

const maxLangs = int(10)

var matchers = map[[maxLangs]string]language.Matcher{}
var matchersMutex sync.RWMutex

func getMatcher(fieldKey string, translations KeyLangValueMap) language.Matcher {
	var langsKey [maxLangs]string
	var i int
	var tMap = translations[fieldKey]

	// Build languages string key
	if _, ok := tMap[defaultLanguageString]; ok {
		langsKey[i] = defaultLanguageString
		i++
	}

	for lang := range tMap {
		if lang != defaultLanguageString {
			langsKey[i] = lang
			i++
		}
	}

	// Return cached matcher for that string key if it's set
	matchersMutex.RLock()
	matcher, ok := matchers[langsKey]
	matchersMutex.RUnlock()

	if ok {
		return matcher
	}

	// Cache missed. Lets create matcher and add it to cache
	langs := make([]language.Tag, 0, i)

	for j := 0; j < i; j++ {
		langs = append(langs, *getTagByString(langsKey[j]))
	}

	matcher = language.NewMatcher(langs)

	matchersMutex.Lock()
	matchers[langsKey] = matcher
	matchersMutex.Unlock()

	return matcher
}

var tags = map[string]language.Tag{}
var tagsMutex sync.RWMutex

func getTagByString(s string) *language.Tag {
	tagsMutex.RLock()
	tag, ok := tags[s]
	tagsMutex.RUnlock()

	if ok {
		return &tag
	}

	tag = language.Make(s)

	tagsMutex.Lock()
	tags[s] = tag
	tagsMutex.Unlock()

	return &tag
}
