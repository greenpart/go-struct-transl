package godatai18n

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"testing"
)

type TrType struct {
	Name         string
	Element      string
	Translations Translations
}

type TrF struct {
	N string
	T string
}

func (t TrF) StructFieldName() string    { return t.N }
func (t TrF) TranslationKeyName() string { return t.T }

type TrTypeTrProvider struct{}

func (t TrTypeTrProvider) TranslatedFields() []TranslatedField {
	return []TranslatedField{
		TrF{"Name", "name"},
		TrF{"Element", "element"},
	}
}

var in = TrType{
	Name:    "",
	Element: "",
	Translations: Translations{
		"en": map[string]string{
			"name":    "John",
			"element": "water",
		},
		"ru": map[string]string{
			"name":    "Джон",
			"element": "вода",
		},
	},
}

func benchmarkTranslator(input TrType, f func(TrType) TrType, b *testing.B) {
	var out TrType
	for i := 0; i < b.N; i++ {
		out = f(input)
	}

	_ = out
}

func noopTranslator(in TrType) (out TrType) {
	return in
}

func fixedLangTranslator(in TrType) TrType {
	out := in
	out.Name = out.Translations["en"]["name"]
	out.Element = out.Translations["en"]["element"]
	return out
}

var enCtx = NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("en")})

func realTranslator(in TrType) TrType {
	TranslateOne(enCtx, &in, TrTypeTrProvider{})
	return in
}

func BenchmarkNoopTranslator(b *testing.B)      { benchmarkTranslator(in, noopTranslator, b) }
func BenchmarkFixedLangTranslator(b *testing.B) { benchmarkTranslator(in, fixedLangTranslator, b) }
func BenchmarkRealTranslator(b *testing.B)      { benchmarkTranslator(in, realTranslator, b) }

func genTrObj() TrType {
	return TrType{
		Name:    "",
		Element: "",
		Translations: Translations{
			"en": map[string]string{
				"name":    "John",
				"element": "water",
			},
			"ru": map[string]string{
				"name":    "Джон",
				"element": "вода",
			},
		},
	}
}

func TestPerfectCase(t *testing.T) {
	enCtx := NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("en")})

	o := genTrObj()
	TranslateOne(enCtx, &o, TrTypeTrProvider{})

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestPerfectCaseWithSecondLang(t *testing.T) {
	ruEnCtx := NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("ru"), language.Make("en")})

	o := genTrObj()
	TranslateOne(ruEnCtx, &o, TrTypeTrProvider{})

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)
}

func TestMissingFirstLang(t *testing.T) {
	jaEnCtx := NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("ja"), language.Make("en")})

	o := genTrObj()
	TranslateOne(jaEnCtx, &o, TrTypeTrProvider{})

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestMissingAllLangsUseEn(t *testing.T) {
	jaPtCtx := NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("ja"), language.Make("pt")})

	o := genTrObj()
	TranslateOne(jaPtCtx, &o, TrTypeTrProvider{})

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestNoLangInContextUseEn(t *testing.T) {
	o := genTrObj()
	TranslateOne(context.Background(), &o, TrTypeTrProvider{})

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestNoEnValuesForDefaltsUsesRu(t *testing.T) {
	o := genTrObj()
	o.Translations["en"] = map[string]string{}
	TranslateOne(context.Background(), &o, TrTypeTrProvider{})

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)
}

func TestNoValues(t *testing.T) {
	o := genTrObj()
	o.Translations = Translations{}
	TranslateOne(context.Background(), &o, TrTypeTrProvider{})

	assert.Equal(t, "", o.Name)
	assert.Equal(t, "", o.Element)
}
