# go-struct-transl [![Build Status](https://travis-ci.org/greenpart/go-struct-transl.svg?branch=master)](https://travis-ci.org/greenpart/go-struct-transl) [![Coverage Status](https://coveralls.io/repos/github/greenpart/go-struct-transl/badge.svg?branch=master)](https://coveralls.io/github/greenpart/go-struct-transl?branch=master)

Translate struct fields and store translations in the same struct.


## Status

Package API is _NOT_ stable yet.

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
	"github.com/greenpart/go-struct-transl/exact"
	"golang.org/x/text/language"
)

type Something struct {
	Name    string `tr:"name"`
	Element string `tr:"element"`
	T       transl.KeyLangValueMap
}

var s = Something{T: transl.KeyLangValueMap{
	"name": map[string]string{
		"en": "John",
		"ru": "Джон",
	},
	"element": map[string]string{
		"en": "water",
	},
},
}

func main() {
	t := exact.NewTranslator()
	t.Translate(&s, []language.Tag{language.Russian})
	fmt.Printf("Name: %s Element: %s\n", s.Name, s.Element)
	// Output: Name: Джон Element: water
}
```

You can see that Name field populated from target `ru` translation and Element field value is from default `en` translation.

In more complex cases with many accepted languages and given translations each field will be set to the best translation available using [golang.org/x/text/language](https://godoc.org/golang.org/x/text/language).


## Accepted languages from HTTP request

You can use `Accept-Language` HTTP header to form language tags. Perhaps
somewhere in your middleware.

``` Go
preferred, _, err := language.ParseAcceptLanguage(request.Header.Get("Accept-Language"))
```

and later

``` Go
t.Translate(&s, preferred)
```


## Default language

You can change default (English) language using `SetDefaults` method of `ExactTranslator`

``` Go
t.SetDefaults("zh", language.Chinese)
```
