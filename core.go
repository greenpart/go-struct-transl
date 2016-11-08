package godatai18n

import (
	"database/sql/driver"
	"encoding/json"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"reflect"
	"sync"
)

// Use struct field with this type to store translations
// e.g. Translations{"en": map[string]string{"name": "John"}}
type Translations map[string]map[string]string

func (m *Translations) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

func (m Translations) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

var matchers = map[string]language.Matcher{}
var matchersMutex sync.RWMutex

func getMatcher(fieldName string, translations Translations) language.Matcher {
	langs := []language.Tag{}
	enFound := false
	for lang, tr := range translations {
		_, ok := tr[fieldName]
		if ok {
			// First language in langs will be fallback option for matcher
			// but map order is not stable,
			// so we need to move en to front, if it's available
			if lang == "en" {
				enFound = true
			} else {
				langs = append(langs, language.Make(lang))
			}
		}
	}
	if enFound {
		langs = append([]language.Tag{language.Make("en")}, langs...)
	}

	langsKey := ""
	for _, lang := range langs {
		langsKey += lang.String()
	}

	matchersMutex.RLock()
	matcher, ok := matchers[langsKey]
	matchersMutex.RUnlock()

	if ok {
		return matcher
	}

	matcher = language.NewMatcher(langs)

	matchersMutex.Lock()
	matchers[langsKey] = matcher
	matchersMutex.Unlock()

	return matcher
}

func translateField(field reflect.Value, fieldName string, translations Translations, targetLanguages []language.Tag) {
	matcher := getMatcher(fieldName, translations)
	effectiveLang, _, _ := matcher.Match(targetLanguages...)
	field.SetString(translations[effectiveLang.String()][fieldName])
}

func TranslateOne(ctx context.Context, target interface{}) {
	meta := metas.getStructMeta(target)
	if len(meta.fields) == 0 {
		return
	}

	structValue := reflect.ValueOf(target)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	translations, ok := structValue.FieldByName("Translations").Interface().(Translations)
	if !ok || len(translations) == 0 {
		return
	}

	targetLanguages, ok := FromContext(ctx)
	if !ok || len(targetLanguages) == 0 {
		targetLanguages = []language.Tag{language.English}
	}

	for _, trF := range meta.fields {
		f := structValue.FieldByName(trF.name)
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			translateField(f, trF.key, translations, targetLanguages)
		}
	}
}

func TranslateMany(ctx context.Context, targets interface{}) {
	v := reflect.ValueOf(targets)

	for i := 0; i < v.Len(); i++ {
		TranslateOne(ctx, v.Index(i).Interface())
	}
}
