package transl

import (
	"fmt"
	"reflect"
	"sync"
)

type fieldMeta struct {
	Name  string
	Key   string
	Index int
}

type StructMeta struct {
	Fields       []fieldMeta
	Translatable bool
	GetterIdx    int

	Err error
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
func buildStructMeta(target interface{}) *StructMeta {
	typ := reflect.TypeOf(target)
	result := StructMeta{
		GetterIdx: -1,
	}

	if typ.Kind() != reflect.Ptr {
		result.Err = fmt.Errorf("Translate of non-pointer type")
		return &result
	}
	t := typ.Elem()

	if _, ok := target.(Translatable); ok {
		result.Translatable = true
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

			result.Fields = append(result.Fields, fm)
		}

		TranslationsGetterType := reflect.TypeOf((*TranslationsGetter)(nil)).Elem()
		if fld.Type.Implements(TranslationsGetterType) {
			result.GetterIdx = i
		}
	}

	if len(result.Fields) == 0 || result.GetterIdx == -1 {
		result.Err = fmt.Errorf("Translate of struct without suitable fields")
		result.Fields = []fieldMeta(nil)
	}

	return &result
}

// Cache metadata building

type structsMetaCache struct {
	metas map[reflect.Type]*StructMeta
	mutex sync.RWMutex
}

var metas = structsMetaCache{map[reflect.Type]*StructMeta{}, sync.RWMutex{}}

func GetStructMeta(target interface{}) (*StructMeta, error) {
	return metas.getStructMeta(target)
}

func (c *structsMetaCache) getStructMeta(target interface{}) (*StructMeta, error) {
	typ := reflect.TypeOf(target)

	c.mutex.RLock()
	meta, ok := c.metas[typ]
	c.mutex.RUnlock()

	if ok {
		return meta, meta.Err
	}

	c.mutex.Lock()
	meta = buildStructMeta(target)
	c.metas[typ] = meta
	c.mutex.Unlock()

	return meta, meta.Err
}
