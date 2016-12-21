package transl

import (
	"reflect"
	"sync"
)

type fieldMeta struct {
	name  string
	key   string
	index int
}

type structMeta struct {
	valid   bool
	fields  []fieldMeta
	trIndex int
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
func buildStructMeta(typ reflect.Type) *structMeta {
	result := structMeta{valid: true, trIndex: -1}

	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
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

		if fld.Name == "Translations" {
			result.trIndex = i
		}
	}

	if len(result.fields) == 0 || result.trIndex == -1 {
		result.valid = false
		result.fields = []fieldMeta{}
	}

	return &result
}

// Cache metadata building

type structsMetaCache struct {
	metas map[reflect.Type]*structMeta
	mutex sync.RWMutex
}

var metas = structsMetaCache{map[reflect.Type]*structMeta{}, sync.RWMutex{}}

func (c *structsMetaCache) getStructMeta(target interface{}) *structMeta {
	typ := indirectType(reflect.TypeOf(target))

	c.mutex.RLock()
	meta, ok := c.metas[typ]
	c.mutex.RUnlock()

	if ok {
		return meta
	}

	c.mutex.Lock()
	meta = buildStructMeta(typ)
	c.metas[typ] = meta
	c.mutex.Unlock()

	return meta
}
