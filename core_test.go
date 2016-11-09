package transl

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"testing"
)

type TrType struct {
	Name         string `tr:"name"`
	Element      string `tr:"element"`
	Translations StringTable
}

var in = TrType{
	Name:    "",
	Element: "",
	Translations: StringTable{
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
	Translate(enCtx, &in)
	return in
}

func BenchmarkNoopTranslator(b *testing.B)      { benchmarkTranslator(in, noopTranslator, b) }
func BenchmarkFixedLangTranslator(b *testing.B) { benchmarkTranslator(in, fixedLangTranslator, b) }
func BenchmarkRealTranslator(b *testing.B)      { benchmarkTranslator(in, realTranslator, b) }

func genTrObj() TrType {
	return TrType{
		Name:    "",
		Element: "",
		Translations: StringTable{
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
	Translate(enCtx, &o)

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestPerfectCaseWithSecondLang(t *testing.T) {
	ruEnCtx := NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("ru"), language.Make("en")})

	o := genTrObj()
	Translate(ruEnCtx, &o)

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)
}

func TestMissingFirstLang(t *testing.T) {
	jaEnCtx := NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("ja"), language.Make("en")})

	o := genTrObj()
	Translate(jaEnCtx, &o)

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestMissingAllLangsUseEn(t *testing.T) {
	jaPtCtx := NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("ja"), language.Make("pt")})

	o := genTrObj()
	Translate(jaPtCtx, &o)

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestNoLangInContextUseEn(t *testing.T) {
	o := genTrObj()
	Translate(context.Background(), &o)

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestNoEnValuesForDefaltsUsesRu(t *testing.T) {
	o := genTrObj()
	o.Translations["en"] = map[string]string{}
	Translate(context.Background(), &o)

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)
}

func TestNoValues(t *testing.T) {
	o := genTrObj()
	o.Translations = StringTable{}
	Translate(context.Background(), &o)

	assert.Equal(t, "", o.Name)
	assert.Equal(t, "", o.Element)
}

func TestOtherDefaults(t *testing.T) {
	SetDefaults("ru", language.Make("ru"))

	o := genTrObj()
	Translate(context.Background(), &o)

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)

	SetDefaults(defaultLanguageString, defaultLanguageTag)
}

// Edge cases for missing/invalid `Translations` and translated field

type NoTranslationsFieldType struct {
	Name string `tr:"name"`
}

func TestNoTranslationsField(t *testing.T) {
	o := NoTranslationsFieldType{}
	Translate(context.Background(), &o)

	assert.Equal(t, "", o.Name)
}

type OtherTranslationsFieldType struct {
	Name         string `tr:"name"`
	Translations int
}

func TestOtherTranslationsField(t *testing.T) {
	o := OtherTranslationsFieldType{}
	Translate(context.Background(), &o)

	assert.Equal(t, "", o.Name)
}

type OtherValueFieldType struct {
	Num          int `tr:"num"`
	Translations StringTable
}

func TestOtherValueField(t *testing.T) {
	o := OtherValueFieldType{}
	Translate(context.Background(), &o)

	assert.Equal(t, 0, o.Num)
}
