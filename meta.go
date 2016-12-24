package transl

import (
	"fmt"
	"reflect"
	"sync"
)

type fieldMeta struct {
	name  string
	key   string
	index int
}

type structMeta struct {
	fields       []fieldMeta
	trIndex      int
	translatable bool
	getterIdx    int

	err error
}

// buildStructMeta fills structMeta with translation-enabled
// fields and returns it
//
// Translation-enabled fields should use struct tags, e.g
//
// type T struct {
// 		First  string `tr:"."`      // key will be equal to field name
// 		Second string `tr:"sec"`	// key is set to `sec`
// }
func buildStructMeta(target interface{}) *structMeta {
	typ := reflect.TypeOf(target)
	result := structMeta{
		trIndex:   -1,
		getterIdx: -1,
	}

	if typ.Kind() != reflect.Ptr {
		result.err = fmt.Errorf("Translate of non-pointer type")
		return &result
	}
	t := typ.Elem()

	if _, ok := target.(Translatable); ok {
		result.translatable = true
		return &result
	}

	for i := 0; i < t.NumField(); i++ {
		fld := t.Field(i)
		tag := fld.Tag.Get("tr")

		if tag != "" {
			name := fld.Name
			key := tag
			if tag == "." {
				key = name
			}

			fm := fieldMeta{name, key, i}

			result.fields = append(result.fields, fm)
		}

		// fmt.Printf("field %+v of %+v type\n", fld.Name, fld.Type)
		TranslationsGetterType := reflect.TypeOf((*TranslationsGetter)(nil)).Elem()
		if fld.Type.Implements(TranslationsGetterType) {
			result.getterIdx = i
		}

		if fld.Name == "Translations" {
			result.trIndex = i
		}
	}

	if len(result.fields) == 0 || result.getterIdx == -1 {
		result.err = fmt.Errorf("Translate of struct without suitable fields")
		result.fields = []fieldMeta(nil)
	}

	return &result
}

// Cache metadata building

type structsMetaCache struct {
	metas map[reflect.Type]*structMeta
	mutex sync.RWMutex
}

var metas = structsMetaCache{map[reflect.Type]*structMeta{}, sync.RWMutex{}}

func (c *structsMetaCache) getStructMeta(target interface{}) (*structMeta, error) {
	typ := reflect.TypeOf(target)

	c.mutex.RLock()
	meta, ok := c.metas[typ]
	c.mutex.RUnlock()

	if ok {
		return meta, meta.err
	}

	c.mutex.Lock()
	meta = buildStructMeta(target)
	c.metas[typ] = meta
	c.mutex.Unlock()

	return meta, meta.err
}
