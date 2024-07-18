// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:generate go run generator.go

package lunes

import (
	"fmt"

	"golang.org/x/text/language"
)

// A Locale provides a collection of time layouts translations to a specific language.
// It is used to provide a map of translations from this locale to English.
type Locale interface {
	// Language represents a BCP 47 language tag, specifying this locale language.
	Language() *language.Tag

	// LongDayNames returns the long day names translations for the week days.
	// It must be sorted, starting from Sunday to Saturday, and contains all 7 elements,
	// even if one or more days are empty. If this locale does not support this format,
	// it should return an empty slice.
	LongDayNames() []string

	// ShortDayNames returns the short day names translations for the week days.
	// It must be sorted, starting from Sunday to Saturday, and contains all 7 elements,
	// even if one or more days are empty. If this locale does not support this format,
	// it should return an empty slice.
	ShortDayNames() []string

	// LongMonthNames returns the long day names translations for the months names.
	// It must be sorted, starting from January to December, and contains all 12 elements,
	// even if one or more months are empty. If this locale does not support this format,
	// it should return an empty slice.
	LongMonthNames() []string

	// ShortMonthNames returns the short day names translations for the months names.
	// It must be sorted, starting from January to December, and contains all 12 elements,
	// even if one or more months are empty. If this locale does not support this format,
	// it should return an empty slice.
	ShortMonthNames() []string

	// DayPeriods returns the periods of day translations for the AM and PM abbreviations.
	// It must be sorted, starting from AM to PM, and contains both elements, even if one
	// of them is empty. If this locale does not support this format, it should return an
	// empty slice.
	DayPeriods() []string
}

type genericLocale struct {
	lang  *language.Tag
	table [5][]string
}

func (g *genericLocale) LongDayNames() []string {
	return g.table[longDayNamesField]
}

func (g *genericLocale) ShortDayNames() []string {
	return g.table[shortDayNamesField]
}

func (g *genericLocale) LongMonthNames() []string {
	return g.table[longMonthNamesField]
}

func (g *genericLocale) ShortMonthNames() []string {
	return g.table[shortMonthNamesField]
}

func (g *genericLocale) DayPeriods() []string {
	return g.table[dayPeriodsField]
}

func (g *genericLocale) Language() *language.Tag {
	return g.lang
}

// ErrUnsupportedLocale indicates that a provided language.Tag is not supported by the
// default CLDR generic locales.
type ErrUnsupportedLocale struct {
	lang *language.Tag
}

func (e *ErrUnsupportedLocale) Error() string {
	return fmt.Sprintf("locale %s not supported", e.lang.String())
}

// NewDefaultLocale creates a new generic locale based on the CLDR gregorian calendar
// translations, and on the provided BCP 47 language.Tag.
//
// If the language.Tag is unknown and no default data is found, it returns
// ErrUnsupportedLocale.
func NewDefaultLocale(lang *language.Tag) (Locale, error) {
	table, ok := tables[lang.String()]
	if !ok {
		return nil, &ErrUnsupportedLocale{lang}
	}

	locale := genericLocale{lang: lang, table: table}
	return &locale, nil
}