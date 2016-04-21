package godatai18n

import (
	"reflect"
)

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
