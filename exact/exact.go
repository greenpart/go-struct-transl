package exact

import (
	"github.com/greenpart/go-struct-transl"
	"golang.org/x/text/language"
	"reflect"
	"sync"
)

// ExactTranslater assigns best suitable translation for each field separately.
// Result may have fields assigned to different language values.
type ExactTranslater interface {
	transl.Translater
	SetDefaults(str string, tag language.Tag)
}

type exactTranslater struct {
	defaultString string
	defaultTag    language.Tag
}

// SetDefaults redefines default language string and tag
func (t *exactTranslater) SetDefaults(str string, tag language.Tag) {
	t.defaultString = str
	t.defaultTag = tag
}

// NewTranslater returns new ExactTranslater
func NewTranslater() ExactTranslater {
	return &exactTranslater{"en", language.English}
}

// Translate applies translation to target.
//
// If target implements Translatable interface
// this function calls Translate method on target.
func (t exactTranslater) Translate(target interface{}, preferred []language.Tag) error {
	meta, err := transl.GetStructMeta(target)
	if err != nil {
		return err
	}

	// Translate target with its Translate method
	// if it implements Translatable interface
	if meta.Translatable {
		tr := target.(transl.Translatable)
		return tr.Translate(preferred)
	}

	return t.translateStructWithGetterField(target, preferred, meta)
}

func (t exactTranslater) translateStructWithGetterField(target interface{}, preferred []language.Tag, meta *transl.StructMeta) error {
	structValue := reflect.Indirect(reflect.ValueOf(target))

	getter := structValue.Field(meta.GetterIdx).Interface().(transl.TranslationsGetter)

	translations := getter.GetTranslations()
	// Empty translations Don't produce error
	if len(translations) == 0 {
		return nil
	}

	if len(preferred) == 0 {
		preferred = []language.Tag{t.defaultTag}
	}

	for _, trF := range meta.Fields {
		f := structValue.Field(trF.Index)
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			t.translateField(f, trF.Key, translations, preferred)
		}
	}

	return nil
}

func (t exactTranslater) translateField(field reflect.Value, fieldKey string, translations transl.KeyLangValueMap, preferred []language.Tag) {
	matcher := t.getMatcher(fieldKey, translations)
	effectiveLang, _, _ := matcher.Match(preferred...)
	field.SetString(translations[fieldKey][effectiveLang.String()])
}

const maxLangs = int(10)

var matchers = map[[maxLangs]string]language.Matcher{}
var matchersMutex sync.RWMutex

func (t *exactTranslater) getMatcher(fieldKey string, translations transl.KeyLangValueMap) language.Matcher {
	var langsKey [maxLangs]string
	var i int
	var tMap = translations[fieldKey]

	// Build languages string key
	if _, ok := tMap[t.defaultString]; ok {
		langsKey[i] = t.defaultString
		i++
	}

	for lang := range tMap {
		if lang != t.defaultString {
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
		langs = append(langs, *t.getTagByString(langsKey[j]))
	}

	matcher = language.NewMatcher(langs)

	matchersMutex.Lock()
	matchers[langsKey] = matcher
	matchersMutex.Unlock()

	return matcher
}

var tags = map[string]language.Tag{}
var tagsMutex sync.RWMutex

func (t *exactTranslater) getTagByString(s string) *language.Tag {
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
