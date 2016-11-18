/*
Package transl translates struct fields and store translations
in the same struct.
*/
package transl

import (
	"database/sql/driver"
	"encoding/json"
	"golang.org/x/net/context"
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

// StringTable is a type for struct field to hold translations
// e.g. Translations{"en": map[string]string{"name": "John"}}
type StringTable map[string]map[string]string

// Scan unmarshals translations from JSON
func (m *StringTable) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

// Value marshals translations to JSON
func (m StringTable) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Translate fills fields of `target` struct with translated values
//
func Translate(ctx context.Context, target interface{}) {
	meta := metas.getStructMeta(target)
	if !meta.valid {
		return
	}

	structValue := reflect.Indirect(reflect.ValueOf(target))

	translations, ok := structValue.Field(meta.trIndex).Interface().(StringTable)
	if !ok || len(translations) == 0 {
		return
	}

	targetLanguages, ok := AcceptedLanguagesFromContext(ctx)
	if !ok || len(targetLanguages) == 0 {
		targetLanguages = []language.Tag{defaultLanguageTag}
	}

	for _, trF := range meta.fields {
		f := structValue.FieldByName(trF.name)
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			translateField(f, trF.key, translations, targetLanguages)
		}
	}
}

func translateField(field reflect.Value, fieldName string, translations StringTable, targetLanguages []language.Tag) {
	matcher := getMatcher(fieldName, translations)
	effectiveLang, _, _ := matcher.Match(targetLanguages...)
	field.SetString(translations[effectiveLang.String()][fieldName])
}

const maxLangs = int(10)

var matchers = map[[maxLangs]string]language.Matcher{}
var matchersMutex sync.RWMutex

func getMatcher(fieldName string, translations StringTable) language.Matcher {
	var langsKey [maxLangs]string
	var i int

	// Build languages string key
	v, ok := translations[defaultLanguageString]
	if ok {
		_, ok = v[fieldName]
		if ok {
			langsKey[i] = defaultLanguageString
			i++
		}
	}

	for lang, tr := range translations {
		_, ok = tr[fieldName]

		if ok {
			if lang == defaultLanguageString {
			} else {
				langsKey[i] = lang
				i++
			}
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
