package godatai18n

import (
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
