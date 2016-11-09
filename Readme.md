# go-struct-transl

Translate struct fields and store translations in the same struct.


## Motivation

Sometimes we have structs those fields might be user-translated to different
languages and we want to store these translations inside structs themselves.

In this case we need a field to hold translations and something to change
other field values according to desired languages.


## Installing

``` Shell
go get github.com/greenpart/go-struct-transl
```


## Usage

Suppose you have a `struct` with fields, that can be translated to different
languages.

Using this package fields can be filled with appropriate translated values.

``` Go
package main

import (
	"fmt"
	"github.com/greenpart/go-struct-transl"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
)

type Something struct {
	Name         string `tr:"name"`
	Element      string `tr:"element"`
	Translations transl.StringTable
}

var s = Something{Translations: transl.StringTable{
	"en": map[string]string{
		"name":    "John",
		"element": "water",
	},
	"ru": map[string]string{
		"name": "Джон",
	},
},
}

var ruCtx = transl.NewContextWithAcceptedLanguages(context.Background(), []language.Tag{language.Russian})

func main() {
	transl.Translate(ruCtx, &s)
	fmt.Printf("Name: %s Element: %s\n", s.Name, s.Element)
	// Name: Джон Element: water
}
```

You can see that Name field populated from target `ru` translation and Element field value is from default `en` translation.

In more complex cases with many accepted languages and given translations each field will be set to the best translation available.


## Accepted languages from HTTP request

You can use `Accept-Language` HTTP header to form language tags. Perhaps
somewhere in your middleware.

``` Go
if tags, _, err := language.ParseAcceptLanguage(request.Header.Get("Accept-Language")); err == nil {
	ctx = transl.NewContextWithAcceptedLanguages(ctx, tags)
}

return ctx
```


## Default language

You can change default (English) language using

``` Go
transl.SetDefaults("zh", language.Chinese)
```
