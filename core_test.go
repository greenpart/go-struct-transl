package transl

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"math/rand"
	"testing"
)

type TrType struct {
	Name         string `tr:"name"`
	Element      string `tr:"element"`
	Translations StringTable
}

var availableLangs = []language.Tag{
	language.English, language.Danish, language.Chinese, language.ModernStandardArabic,
	language.BrazilianPortuguese, language.Swahili, language.SimplifiedChinese,
	language.Russian, language.Norwegian, language.Turkish, language.Urdu}

func getRandLang() *language.Tag {
	return &availableLangs[rand.Intn(len(availableLangs))]
}

var contexts []context.Context
var currentContext context.Context

var enCtx = NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Make("en")})

func oldBenchmarkTranslator(input TrType, f func(TrType) TrType, b *testing.B) {
	var out TrType
	currentContext = enCtx

	for i := 0; i < b.N; i++ {
		out = f(input)
	}

	_ = out
}

func benchmarkTranslator(f func(TrType) TrType, b *testing.B) {
	inputsCount := 10000
	inputs := []TrType{}
	for i := 0; i < inputsCount; i++ {
		in := TrType{Translations: StringTable{}}
		for j := 0; j < 3+rand.Intn(12); j++ { // 3-15 translations
			l := getRandLang()

			if rand.Intn(3) > 0 { // Every 2 of 3 for this field
				if _, ok := in.Translations[l.String()]; !ok {
					in.Translations[l.String()] = map[string]string{}
				}
				in.Translations[l.String()]["name"] = fmt.Sprintf("%+v", rand.Float64())
			}

			if rand.Intn(3) > 0 { // Every 2 of 3
				if _, ok := in.Translations[l.String()]; !ok {
					in.Translations[l.String()] = map[string]string{}
				}
				in.Translations[l.String()]["element"] = fmt.Sprintf("%+v", rand.Float64())
			}
		}
		inputs = append(inputs, in)
	}

	contextsCount := 10000
	for i := 0; i < contextsCount; i++ {
		langs := []language.Tag{}

		for j := 0; j < 2+rand.Intn(5); j++ {
			l := getRandLang()

			if rand.Intn(3) > 0 { // Every 2 of 3
				langs = append(langs, *l)
			}
		}

		contexts = append(contexts, NewContextWithAcceptedLanguages(context.Background(), langs))
	}

	var out TrType

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentContext = contexts[i%contextsCount]
		out = f(inputs[i%inputsCount])
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

func realTranslator(in TrType) TrType {
	Translate(currentContext, &in)
	return in
}

func BenchmarkNoopTranslator(b *testing.B)      { benchmarkTranslator(noopTranslator, b) }
func BenchmarkFixedLangTranslator(b *testing.B) { benchmarkTranslator(fixedLangTranslator, b) }
func BenchmarkRealTranslator(b *testing.B)      { oldBenchmarkTranslator(genTrObj(), realTranslator, b) }
func BenchmarkReRandTranslator(b *testing.B)    { benchmarkTranslator(realTranslator, b) }

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

func TestStringTableJsonScan(t *testing.T) {
	st := StringTable{}
	err := st.Scan([]uint8(`{}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, StringTable{}, st)

	st = StringTable{}
	err = st.Scan([]uint8(`{"en":{}}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, StringTable{"en": map[string]string{}}, st)

	st = StringTable{}
	err = st.Scan([]uint8(`{"en":{"name":"Bob"}}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, StringTable{"en": map[string]string{"name": "Bob"}}, st)

	st = StringTable{}
	err = st.Scan([]uint8(`{"en":{"name":"Bob"}, "ru":{"element": "earth"}}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, StringTable{"en": map[string]string{"name": "Bob"}, "ru": map[string]string{"element": "earth"}}, st)
}

func TestStringTableJsonValue(t *testing.T) {
	st := StringTable{}
	v, err := st.Value()
	assert.Equal(t, nil, err)
	assert.Equal(t, "{}", v)

	st = StringTable{"en": map[string]string{}}
	v, err = st.Value()
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"en":{}}`, v)
}
