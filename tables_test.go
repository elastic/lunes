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

// These tests are not meant to validate the whole tables file neither the translations.
// instead, it checks well-known conditions such as language values overrides and the
// tables.go file format.
package lunes

import (
	"fmt"
	"slices"
	"testing"
)

func TestLocaleTableEn(t *testing.T) {
	localeStdTable := [5][]string{
		shortDayNamesStd,
		longDayNamesStd,
		shortMonthNamesStd,
		longMonthNamesStd,
		dayPeriodsStd,
	}

	lang := "en"
	testTableEquality(t, lang, localeTableEn, localeStdTable, shortMonthNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEn, localeStdTable, longDayNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEn, localeStdTable, shortMonthNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEn, localeStdTable, longMonthNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEn, localeStdTable, dayPeriodsField, 0, 0)
}

func TestLocaleTableEn001Overrides(t *testing.T) {
	lang := "en-0001"

	// "en-001" inherits "en" overriding the short month names (Sep to Sept) and day periods to lowercase
	testTableEquality(t, lang, localeTableEn001, localeTableEn, shortDayNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEn001, localeTableEn, longDayNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEn001, localeTableEn, shortMonthNamesField, 0, 8)

	if localeTableEn001[shortMonthNamesField][8] != "Sept" {
		t.Errorf("'%s' expected shortMonthNamesField[8] value: Sept, got: %v", lang, localeTableEn001[shortMonthNamesField][8])
	}

	testTableEquality(t, lang, localeTableEn001, localeTableEn, shortMonthNamesField, 9, 0)
	testTableEquality(t, lang, localeTableEn001, localeTableEn, longMonthNamesField, 0, 0)

	if localeTableEn001[dayPeriodsField][0] != "am" {
		t.Errorf("'%s' expected dayPeriodsField[0] value: am, got: %v", lang, localeTableEn001[dayPeriodsField][0])
	}

	if localeTableEn001[dayPeriodsField][1] != "pm" {
		t.Errorf("'%s' expected dayPeriodsField[1] value: pm, got: %v", lang, localeTableEn001[dayPeriodsField][1])
	}
}

func TestLocaleTableEnAuOverrides(t *testing.T) {
	lang := "en-AU"

	// "en-AU" inherits "en-001" overriding the short month name (June July Aug Sept) and day periods to lowercase
	testTableEquality(t, lang, localeTableEnAU, localeTableEn001, shortDayNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEnAU, localeTableEn001, longDayNamesField, 0, 0)
	testTableEquality(t, lang, localeTableEnAU, localeTableEn001, shortMonthNamesField, 0, 5)

	if localeTableEnAU[shortMonthNamesField][5] != "June" {
		t.Errorf("'%s' expected shortMonthNamesField][5] value: June, got: %v", lang, localeTableEnAU[shortMonthNamesField][5])
	}

	if localeTableEnAU[shortMonthNamesField][6] != "July" {
		t.Errorf("'%s' expected shortMonthNamesField][6] value: July, got: %v", lang, localeTableEnAU[shortMonthNamesField][6])
	}

	testTableEquality(t, lang, localeTableEnAU, localeTableEn001, shortMonthNamesField, 7, 8)

	if localeTableEnAU[shortMonthNamesField][8] != "Sept" {
		t.Errorf("'%s' expected shortMonthNamesField][8] value: Sept, got: %v", lang, localeTableEnAU[shortMonthNamesField][8])
	}

	testTableEquality(t, lang, localeTableEnAU, localeTableEn001, shortMonthNamesField, 9, 0)
	testTableEquality(t, lang, localeTableEnAU, localeTableEn001, longMonthNamesField, 0, 0)

	if localeTableEnAU[dayPeriodsField][0] != "am" {
		t.Errorf("'%s' expected dayPeriodsField[0] value: am, got: %v", lang, localeTableEnAU[dayPeriodsField][0])
	}

	if localeTableEnAU[dayPeriodsField][1] != "pm" {
		t.Errorf("'%s' expected dayPeriodsField[1] value: pm, got: %v", lang, localeTableEnAU[dayPeriodsField][1])
	}
}

func testTableEquality(t *testing.T, lang string, value, expected [5][]string, field, from, to int) {
	var fieldName string
	switch field {
	case shortDayNamesField:
		fieldName = "shortDayNamesField"
	case longDayNamesField:
		fieldName = "longDayNamesField"
	case shortMonthNamesField:
		fieldName = "shortMonthNamesField"
	case longMonthNamesField:
		fieldName = "longMonthNamesField"
	case dayPeriodsField:
		fieldName = "dayPeriodsField"
	}

	if from != 0 || to != 0 {
		fieldName += fmt.Sprintf("[%d:%d]", from, to)
	}

	if len(value[field]) != len(expected[field]) {
		t.Errorf("'%s' expected %s value: %s, got: %v", lang, fieldName, value, expected)
		return
	}

	if to == 0 {
		to = len(expected[field])
	}

	if !slices.Equal(value[field][from:to], expected[field][from:to]) {
		t.Errorf("'%s' expected %s value: %s, got: %v", lang, fieldName, value[field][from:to], expected[field][from:to])
	}
}

func TestConstFieldsOrder(t *testing.T) {
	t.Run("shortDayNamesField", func(t *testing.T) {
		if shortDayNamesField != 0 {
			t.Errorf("shortDayNamesField value must be 0")
		}
	})

	t.Run("longDayNamesField", func(t *testing.T) {
		if longDayNamesField != 1 {
			t.Errorf("longDayNamesField value must be 1")
		}
	})

	t.Run("longDayNamesField", func(t *testing.T) {
		if shortMonthNamesField != 2 {
			t.Errorf("longDayNamesField value must be 2")
		}
	})

	t.Run("longMonthNamesField", func(t *testing.T) {
		if longMonthNamesField != 3 {
			t.Errorf("longMonthNamesField value must be 3")
		}
	})

	t.Run("dayPeriodsField", func(t *testing.T) {
		if dayPeriodsField != 4 {
			t.Errorf("dayPeriodsField value must be 4")
		}
	})
}
