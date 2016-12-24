package transl_test

import (
	"fmt"
	"github.com/greenpart/go-struct-transl"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"math/rand"
	"testing"
)

type TrType struct {
	Name         string `tr:"name"`
	Element      string `tr:"element"`
	Translations transl.KeyLangValueMap
}

var availableLangs = []language.Tag{
	language.English, language.Danish, language.Chinese, language.ModernStandardArabic,
	language.BrazilianPortuguese, language.Swahili, language.SimplifiedChinese,
	language.Russian, language.Norwegian, language.Turkish, language.Urdu}

func getRandLang() *language.Tag {
	return &availableLangs[rand.Intn(len(availableLangs))]
}

var preferredSlices [][]language.Tag
var currentPreferred []language.Tag

var enPreferred = []language.Tag{language.English}

func oldBenchmarkTranslator(input TrType, f func(TrType) TrType, b *testing.B) {
	var out TrType
	currentPreferred = enPreferred

	for i := 0; i < b.N; i++ {
		out = f(input)
	}

	_ = out
}

func benchmarkTranslator(f func(TrType) TrType, b *testing.B) {
	inputsCount := 10000
	inputs := []TrType{}
	for i := 0; i < inputsCount; i++ {
		in := TrType{Translations: transl.KeyLangValueMap{
			"name":    map[string]string{},
			"element": map[string]string{},
		}}
		for j := 0; j < 3+rand.Intn(12); j++ { // 3-15 translations
			l := getRandLang()

			if rand.Intn(3) > 0 { // Every 2 of 3 for this field
				in.Translations["name"][l.String()] = fmt.Sprintf("%+v", rand.Float64())
			}

			if rand.Intn(3) > 0 { // Every 2 of 3
				in.Translations["element"][l.String()] = fmt.Sprintf("%+v", rand.Float64())
			}
		}
		inputs = append(inputs, in)
	}

	preferredsCount := 10000
	for i := 0; i < preferredsCount; i++ {
		langs := []language.Tag{}

		for j := 0; j < 2+rand.Intn(5); j++ {
			l := getRandLang()

			if rand.Intn(3) > 0 { // Every 2 of 3
				langs = append(langs, *l)
			}
		}

		preferredSlices = append(preferredSlices, langs)
	}

	var out TrType

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentPreferred = preferredSlices[i%preferredsCount]
		out = f(inputs[i%inputsCount])
	}

	_ = out
}

func noopTranslator(in TrType) (out TrType) {
	return in
}

func fixedLangTranslator(in TrType) TrType {
	out := in
	out.Name = out.Translations["name"]["en"]
	out.Element = out.Translations["element"]["en"]
	return out
}

func realTranslator(in TrType) TrType {
	transl.Translate(&in, currentPreferred)
	return in
}

func BenchmarkNoopTranslator(b *testing.B)      { benchmarkTranslator(noopTranslator, b) }
func BenchmarkFixedLangTranslator(b *testing.B) { benchmarkTranslator(fixedLangTranslator, b) }
func BenchmarkRealTranslator(b *testing.B)      { oldBenchmarkTranslator(genTrObj(), realTranslator, b) }
func BenchmarkReRandTranslator(b *testing.B)    { benchmarkTranslator(realTranslator, b) }

type TS struct {
	called int
}

func (ts *TS) Translate(preferred []language.Tag) error {
	ts.called += 1
	return nil
}

func TestTranslateNonPointer(t *testing.T) {
	ts := TS{}
	err := transl.Translate(ts, []language.Tag{})

	if assert.NotNil(t, err) {
		assert.Equal(t, fmt.Errorf("Translate of non-pointer type"), err)
	}
}

func TestTranslateTranslatable(t *testing.T) {
	ts := TS{}
	err := transl.Translate(&ts, []language.Tag{})

	assert.Nil(t, err)
	assert.Equal(t, 1, ts.called)
}

func genTrObj() TrType {
	return TrType{
		Name:    "",
		Element: "",
		Translations: transl.KeyLangValueMap{
			"name": map[string]string{
				"en": "John",
				"ru": "Джон",
			},
			"element": map[string]string{
				"en": "water",
				"ru": "вода",
			},
		},
	}
}

func TestPerfectCase(t *testing.T) {
	s := genTrObj()
	transl.Translate(&s, []language.Tag{language.English})

	assert.Equal(t, "John", s.Name)
	assert.Equal(t, "water", s.Element)
}

func TestPerfectCaseWithSecondLang(t *testing.T) {
	o := genTrObj()
	transl.Translate(&o, []language.Tag{language.Russian, language.English})

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)
}

func TestMissingFirstLang(t *testing.T) {
	o := genTrObj()
	transl.Translate(&o, []language.Tag{language.Japanese, language.English})

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestMissingAllLangsUseEn(t *testing.T) {
	o := genTrObj()
	transl.Translate(&o, []language.Tag{language.Japanese, language.Portuguese})

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestNoPreferredLang(t *testing.T) {
	o := genTrObj()
	transl.Translate(&o, []language.Tag{})

	assert.Equal(t, "John", o.Name)
	assert.Equal(t, "water", o.Element)
}

func TestNoEnValuesForDefaltsUsesRu(t *testing.T) {
	o := genTrObj()
	delete(o.Translations["name"], "en")
	delete(o.Translations["element"], "en")
	transl.Translate(&o, []language.Tag{})

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)
}

func TestNoValues(t *testing.T) {
	o := genTrObj()
	o.Translations = transl.KeyLangValueMap{}
	transl.Translate(&o, []language.Tag{})

	assert.Equal(t, "", o.Name)
	assert.Equal(t, "", o.Element)
}

func TestOtherDefaults(t *testing.T) {
	transl.SetDefaults("ru", language.Russian)

	o := genTrObj()
	transl.Translate(&o, []language.Tag{})

	assert.Equal(t, "Джон", o.Name)
	assert.Equal(t, "вода", o.Element)

	transl.SetDefaults("en", language.English)
}

// Edge cases for missing/invalid `Translations` and translated field

func TestNoTranslationsField(t *testing.T) {
	type T struct {
		Name string
	}

	o := T{}
	transl.Translate(&o, []language.Tag{})

	assert.Equal(t, "", o.Name)
}

func TestOtherTranslationsField(t *testing.T) {
	type T struct {
		Name         string `tr:"name"`
		Translations int
	}

	o := T{}
	transl.Translate(&o, []language.Tag{})

	assert.Equal(t, "", o.Name)
}

func TestOtherValueField(t *testing.T) {
	type T struct {
		Num          int `tr:"num"`
		Translations transl.KeyLangValueMap
	}

	o := T{}
	transl.Translate(&o, []language.Tag{})

	assert.Equal(t, 0, o.Num)
}

func TestStringTableJsonScan(t *testing.T) {
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

func TestStringTableJsonValue(t *testing.T) {
	st := transl.KeyLangValueMap{}
	v, err := st.Value()
	assert.Equal(t, nil, err)
	assert.Equal(t, "{}", v)

	st = transl.KeyLangValueMap{"name": map[string]string{}}
	v, err = st.Value()
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"name":{}}`, v)
}
