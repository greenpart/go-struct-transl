package godatai18n

import (
	"golang.org/x/net/context"
	"golang.org/x/text/language"
)

type contextKey int

const (
	acceptedLanguagesKey contextKey = iota
)

// NewContext returns context with accepted languages information
// langs is a slice of language.Tag
func NewContext(ctx context.Context, langs []language.Tag) context.Context {
	return context.WithValue(ctx, acceptedLanguagesKey, langs)
}

// FromContext returns accepted languages from context
func FromContext(ctx context.Context) ([]language.Tag, bool) {
	langs, ok := ctx.Value(acceptedLanguagesKey).([]language.Tag)
	return langs, ok
}
