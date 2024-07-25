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

package lunes

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

type ParseTestLocale struct {
	lang        string
	value       string
	expectedErr error
}

type ParseTest struct {
	name     string
	format   string
	stdValue string
	hasTZ    bool // has time zone
	hasWD    bool // has a weekday
	yearSign int  // -1 indicates the year is not present in the layout
	locales  *[]ParseTestLocale
}

//nolint:varcheck // order matters
const (
	sun = iota
	mon
	tue
	wed
	thu
	fri
	sat
)

//nolint:varcheck // order matters
const (
	jan = iota
	feb
	mar
	apr
	may
	jun
	jul
	aug
	sep
	oct
	nov
	dec
)

const (
	am = iota
	pm = iota
)

type replacement struct {
	field int
	value int
}

var emptyReplacements []replacement

func allLocalesTests(valuePattern string, replacements []replacement) *[]ParseTestLocale {
	return genAllLocalesTests(valuePattern, replacements, func(v string) string { return v })
}

func genAllLocalesTests(valuePattern string, replacements []replacement, formatter func(v string) string) *[]ParseTestLocale {
	var tests []ParseTestLocale
	for lang := range tables {
		test := ParseTestLocale{lang: lang}
		locale := genericLocale{lang: lang, table: tables[lang]}

		args := make([]any, 0, len(replacements)/2)
		var unsupportedElem string
		for _, repl := range replacements {
			fieldVal := getLocaleFieldValue(&locale, repl.field)
			if len(fieldVal) == 0 {
				// fallback to english so it replaces the placeholders with any value
				fieldVal = tables["en"][repl.field]
				if unsupportedElem == "" {
					switch repl.field {
					case shortDayNamesField, longDayNamesField:
						unsupportedElem = fieldVal[mon]
					case shortMonthNamesField, longMonthNamesField:
						unsupportedElem = fieldVal[jan]
					case dayPeriodsField:
						unsupportedElem = "PM"
					}
				}
			}

			var fv string
			if formatter != nil {
				fv = formatter(fieldVal[repl.value])
			} else {
				fv = fieldVal[repl.value]
			}

			args = append(args, fv)
		}

		if unsupportedElem != "" {
			test.expectedErr = newUnsupportedLayoutElemError(unsupportedElem, &locale)
		}

		test.value = fmt.Sprintf(valuePattern, args...)
		tests = append(tests, test)
	}

	return &tests
}

func getLocaleFieldValue(locale Locale, field int) []string {
	switch field {
	case shortDayNamesField:
		return locale.ShortDayNames()
	case longDayNamesField:
		return locale.LongDayNames()
	case shortMonthNamesField:
		return locale.ShortMonthNames()
	case longMonthNamesField:
		return locale.LongMonthNames()
	case dayPeriodsField:
		return locale.DayPeriods()
	default:
		return nil
	}
}

// Most of these tests were taken from the time.format_test.go file.
// Even if it's not translating any placeholder, it should keep working and don't break the underline
// time.Parse functionality.
var parseTests = []ParseTest{
	{
		name:     "ANSIC",
		format:   time.ANSIC,
		stdValue: "Thu May  1 21:00:57 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  1 21:00:57 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, may}})},
	{
		name:     "UnixDate",
		format:   time.UnixDate,
		stdValue: "Fri Sep  4 21:00:57 PST 2010",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57 PST 2010", []replacement{{shortDayNamesField, fri}, {shortMonthNamesField, sep}}),
	},
	{
		name:     "RubyDate",
		format:   time.RubyDate,
		stdValue: "Tue Dec 05 21:00:57 -0800 2010",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s 05 21:00:57 -0800 2010", []replacement{{shortDayNamesField, tue}, {shortMonthNamesField, dec}}),
	},
	{
		name:     "RFC850",
		format:   time.RFC850,
		stdValue: "Tuesday, 04-Dec-10 21:00:57 PST",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s, 04-%s-10 21:00:57 PST", []replacement{{longDayNamesField, tue}, {shortMonthNamesField, dec}}),
	},
	{
		name:     "RFC1123",
		format:   time.RFC1123,
		stdValue: "Thu, 04 Feb 2010 21:00:57 PST",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s, 04 %s 2010 21:00:57 PST", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RFC1123",
		format:   time.RFC1123,
		stdValue: "Thu, 04 Feb 2010 22:00:57 PDT",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s, 04 %s 2010 22:00:57 PDT", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RFC1123Z",
		format:   time.RFC1123Z,
		stdValue: "Thu, 04 Feb 2010 21:00:57 -0800",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s, 04 %s 2010 21:00:57 -0800", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RFC3339",
		format:   time.RFC3339,
		stdValue: "2010-02-04T21:00:57-08:00",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04T21:00:57-08:00", emptyReplacements),
	},
	{
		name:     "custom: \"2006-01-02 15:04:05-07\"",
		format:   "2006-01-02 15:04:05-07",
		stdValue: "2010-02-04 21:00:57-08",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57-08", emptyReplacements),
	},
	// Optional fractional seconds
	{
		name:     "ANSIC",
		format:   time.ANSIC,
		stdValue: "Thu Feb  4 21:00:57.0 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.0 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "UnixDate",
		format:   time.UnixDate,
		stdValue: "Thu Feb  4 21:00:57.01 PST 2010",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.01 PST 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RubyDate",
		format:   time.RubyDate,
		stdValue: "Thu Feb 04 21:00:57.012 -0800 2010",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s 04 21:00:57.012 -0800 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RFC850",
		format:   time.RFC850,
		stdValue: "Thursday, 04-Feb-10 21:00:57.0123 PST",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s, 04-%s-10 21:00:57.0123 PST", []replacement{{longDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RFC1123",
		format:   time.RFC1123,
		stdValue: "Thu, 04 Feb 2010 21:00:57.01234 PST",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s, 04 %s 2010 21:00:57.01234 PST", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RFC1123Z",
		format:   time.RFC1123Z,
		stdValue: "Thu, 04 Feb 2010 21:00:57.01234 -0800",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s, 04 %s 2010 21:00:57.01234 -0800", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "RFC3339",
		format:   time.RFC3339,
		stdValue: "2010-02-04T21:00:57.012345678-08:00",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04T21:00:57.012345678-08:00", emptyReplacements),
	},
	{
		name:     "custom: \"2006-01-02 15:04:05\"",
		format:   "2006-01-02 15:04:05",
		stdValue: "2010-02-04 21:00:57.0",
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.0", emptyReplacements),
	},
	// Amount of white space should not matter.
	{
		name:     "ANSIC",
		format:   time.ANSIC,
		stdValue: "Thu Feb 4 21:00:57 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s 4 21:00:57 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "ANSIC",
		format:   time.ANSIC,
		stdValue: "Thu      Feb     4     21:00:57     2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s      %s     4     21:00:57     2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	// Case should not matter
	{
		name:     "ANSIC",
		format:   time.ANSIC,
		stdValue: "THU FEB 4 21:00:57 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  genAllLocalesTests("%s %s 4 21:00:57 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}, strings.ToUpper),
	},
	{
		name:     "ANSIC",
		format:   time.ANSIC,
		stdValue: "thu feb 4 21:00:57 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  genAllLocalesTests("%s %s 4 21:00:57 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}, strings.ToLower),
	},
	// Fractional seconds
	{
		name:     "millisecond:: dot separator",
		format:   "Mon Jan _2 15:04:05.000 2006",
		stdValue: "Thu Feb  4 21:00:57.012 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.012 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "microsecond:: dot separator",
		format:   "Mon Jan _2 15:04:05.000000 2006",
		stdValue: "Thu Feb  4 21:00:57.012345 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.012345 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "nanosecond:: dot separator",
		format:   "Mon Jan _2 15:04:05.000000000 2006",
		stdValue: "Thu Feb  4 21:00:57.012345678 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.012345678 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "millisecond:: comma separator",
		format:   "Mon Jan _2 15:04:05,000 2006",
		stdValue: "Thu Feb  4 21:00:57.012 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.012 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "microsecond:: comma separator",
		format:   "Mon Jan _2 15:04:05,000000 2006",
		stdValue: "Thu Feb  4 21:00:57.012345 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.012345 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	{
		name:     "nanosecond:: comma separator",
		format:   "Mon Jan _2 15:04:05,000000000 2006",
		stdValue: "Thu Feb  4 21:00:57.012345678 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  4 21:00:57.012345678 2010", []replacement{{shortDayNamesField, thu}, {shortMonthNamesField, feb}}),
	},
	// Leading zeros in other places should not be taken as fractional seconds.
	{
		name:     "zero1",
		format:   "2006.01.02.15.04.05.0",
		stdValue: "2010.02.04.21.00.57.0",
		yearSign: 1,
		locales:  allLocalesTests("2010.02.04.21.00.57.0", emptyReplacements),
	},
	{
		name:     "zero2",
		format:   "2006.01.02.15.04.05.00",
		stdValue: "2010.02.04.21.00.57.01",
		yearSign: 1,
		locales:  allLocalesTests("2010.02.04.21.00.57.01", emptyReplacements),
	},
	// Month and day names only match when not followed by a lower-case letter.
	{
		name:     "Janet",
		format:   "Hi Janet, the Month is January: Jan _2 15:04:05 2006",
		stdValue: "Hi Janet, the Month is February: Feb  4 21:00:57 2010",
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("Hi Janet, the Month is %s: %s  4 21:00:57 2010", []replacement{{longMonthNamesField, feb}, {shortMonthNamesField, feb}}),
	},
	// GMT with offset.
	{
		name:     "GMT-8",
		format:   time.UnixDate,
		stdValue: "Fri Feb  5 05:00:57 GMT-8 2010",
		hasTZ:    true,
		hasWD:    true,
		yearSign: 1,
		locales:  allLocalesTests("%s %s  5 05:00:57 GMT-8 2010", []replacement{{shortDayNamesField, fri}, {shortMonthNamesField, feb}}),
	},
	// Accept any number of fractional second digits (including none)
	{
		format:   "2006-01-02 15:04:05.9999 -0700 MST",
		stdValue: "2010-02-04 21:00:57 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05.999999999 -0700 MST",
		stdValue: "2010-02-04 21:00:57 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05.9999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.0123 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.0123 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05.999999999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.0123 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.0123 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05.9999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.012345678 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.012345678 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05.999999999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.012345678 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.012345678 -0800 PST", emptyReplacements),
	},
	// Comma "," separator.
	{
		format:   "2006-01-02 15:04:05,9999 -0700 MST",
		stdValue: "2010-02-04 21:00:57 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05,999999999 -0700 MST",
		stdValue: "2010-02-04 21:00:57 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05,9999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.0123 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.0123 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05,999999999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.0123 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.0123 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05,9999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.012345678 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.012345678 -0800 PST", emptyReplacements),
	},
	{
		format:   "2006-01-02 15:04:05,999999999 -0700 MST",
		stdValue: "2010-02-04 21:00:57.012345678 -0800 PST",
		hasTZ:    true,
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 21:00:57.012345678 -0800 PST", emptyReplacements),
	},
	{
		format:   time.StampNano,
		stdValue: "Feb  4 21:00:57.012345678",
		yearSign: -1,
		locales:  allLocalesTests("%s  4 21:00:57.012345678", []replacement{{shortMonthNamesField, feb}}),
	},
	{
		format:   "Jan _2 15:04:05.999",
		stdValue: "Feb  4 21:00:57.012300000",
		yearSign: -1,
		locales:  allLocalesTests("%s  4 21:00:57.012300000", []replacement{{shortMonthNamesField, feb}}),
	},
	{
		format:   "Jan _2 15:04:05.999",
		stdValue: "Feb  4 21:00:57.012345678",
		yearSign: -1,
		locales:  allLocalesTests("%s  4 21:00:57.012345678", []replacement{{shortMonthNamesField, feb}}),
	},
	{
		format:   "Jan _2 15:04:05.999999999",
		stdValue: "Feb  4 21:00:57.0123",
		yearSign: -1,
		locales:  allLocalesTests("%s  4 21:00:57.0123", []replacement{{shortMonthNamesField, feb}}),
	},
	{
		format:   "Jan _2 15:04:05.999999999",
		stdValue: "Feb  4 21:00:57.012345678",
		yearSign: -1,
		locales:  allLocalesTests("%s  4 21:00:57.012345678", []replacement{{shortMonthNamesField, feb}}),
	},
	// Day of the year.
	{
		format:   "2006-01-02 002 15:04:05",
		stdValue: "2010-02-04 035 21:00:57",
		yearSign: 1,
		locales:  allLocalesTests("2010-02-04 035 21:00:57", emptyReplacements),
	},
	{
		format:   "2006-01 002 15:04:05",
		stdValue: "2010-02 035 21:00:57",
		yearSign: 1,
		locales:  allLocalesTests("2010-02 035 21:00:57", emptyReplacements),
	},
	{
		format:   "2006-002 15:04:05",
		stdValue: "2010-035 21:00:57",
		yearSign: 1,
		locales:  allLocalesTests("2010-035 21:00:57", emptyReplacements),
	},
	{
		format:   "200600201 15:04:05",
		stdValue: "201003502 21:00:57",
		yearSign: 1,
		locales:  allLocalesTests("201003502 21:00:57", emptyReplacements),
	},
	{
		format:   "200600204 15:04:05",
		stdValue: "201003504 21:00:57",
		yearSign: 1,
		locales:  allLocalesTests("201003504 21:00:57", emptyReplacements),
	},
	// Seconds time zone
	{
		format:   "2006-01-02T15:04:05-070000",
		stdValue: "1871-01-01T05:33:02-003408",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02-003408", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05-07:00:00",
		stdValue: "1871-01-01T05:33:02-00:34:08",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02-00:34:08", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05-070000",
		stdValue: "1871-01-01T05:33:02+003408",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02+003408", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05-07:00:00",
		stdValue: "1871-01-01T05:33:02+00:34:08",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02+00:34:08", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05Z070000",
		stdValue: "1871-01-01T05:33:02-003408",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02-003408", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05Z07:00:00",
		stdValue: "1871-01-01T05:33:02+00:34:08",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02+00:34:08", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05-07",
		stdValue: "1871-01-01T05:33:02+01",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02+01", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05-07",
		stdValue: "1871-01-01T05:33:02-02",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02-02", emptyReplacements),
	},
	{
		format:   "2006-01-02T15:04:05Z07",
		stdValue: "1871-01-01T05:33:02-02",
		yearSign: 1,
		hasTZ:    true,
		locales:  allLocalesTests("1871-01-01T05:33:02-02", emptyReplacements),
	},
	// Underscore 2xxx
	{
		format:   "15:04_20060102",
		stdValue: "14:38_20150618",
		yearSign: -1,
		locales:  allLocalesTests("14:38_20150618", emptyReplacements),
	},
	// Extra layouts
	{
		format:   "Monday Jan 2 2006",
		stdValue: "Wednesday Oct 9 2024",
		yearSign: 1,
		hasWD:    true,
		locales:  allLocalesTests("%s %s 9 2024", []replacement{{longDayNamesField, wed}, {shortMonthNamesField, oct}}),
	},
	{
		format:   "Monday 02 January 2006 15:04:05",
		stdValue: "Wednesday 09 October 2024 21:20:49",
		yearSign: 1,
		hasWD:    true,
		locales:  allLocalesTests("%s 09 %s 2024 21:20:49", []replacement{{longDayNamesField, wed}, {longMonthNamesField, oct}}),
	},
	{
		format:   "Mon January __2 2006 03:04:05PM",
		stdValue: "Thu February 54 2006 09:15:05AM",
		yearSign: 1,
		hasWD:    true,
		locales:  allLocalesTests("%s %s 54 2006 09:15:05%s", []replacement{{shortDayNamesField, thu}, {longMonthNamesField, feb}, {dayPeriodsField, am}}),
	},
	{
		format:   "Mon January 2 2006 03:04:05PM",
		stdValue: "Thu February 4 2006 09:15:05PM",
		yearSign: 1,
		hasWD:    true,
		locales:  allLocalesTests("%s %s 4 2006 09:15:05%s", []replacement{{shortDayNamesField, thu}, {longMonthNamesField, feb}, {dayPeriodsField, pm}}),
	},
}

var defaultLocation *time.Location

func init() {
	var err error
	defaultLocation, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal(err)
	}
}

func TestParseInLocation(t *testing.T) {
	testParseFunc(t, parseTests, func(format, stdValue string) (time.Time, error) {
		return time.ParseInLocation(format, stdValue, defaultLocation)
	}, func(format string, value string, locale string) (time.Time, error) {
		return ParseInLocation(format, value, locale, defaultLocation)
	})
}

func TestParse(t *testing.T) {
	testParseFunc(t, parseTests, time.Parse, Parse)
}

type stdParseFunction func(format, value string) (time.Time, error)

type parseFunction func(format string, value string, locale string) (time.Time, error)

func testParseFunc(t *testing.T, tests []ParseTest, stdFn stdParseFunction, parseFn parseFunction) {
	var err error
	var result, expected time.Time

	for _, test := range tests {
		var name string
		if test.name != "" {
			name = test.name
		} else {
			name = test.format
		}

		t.Run(name, func(t *testing.T) {
			expected, err = stdFn(test.format, test.stdValue)
			if err != nil {
				t.Errorf("%s error: failed to parse stdValue %s: %v", test.name, test.stdValue, err)
			}

			result, err = parseFn(test.format, test.stdValue, LocaleEn)
			if err != nil {
				t.Errorf("%s error: %v", test.name, err)
			} else {
				checkParsedTime(result, expected, test, t)
			}

			if test.locales != nil {
				for _, lt := range *test.locales {
					t.Run(lt.lang, func(t *testing.T) {
						result, err = parseFn(test.format, lt.value, lt.lang)
						if err != nil {
							if lt.expectedErr == nil {
								t.Errorf("%s error parsing '%v': %v", test.name, lt.value, err)
							} else if err.Error() != lt.expectedErr.Error() {
								t.Errorf("%s expected error: '%v', got: '%v'", test.name, lt.expectedErr, err)
							}
							return
						} else if lt.expectedErr != nil {
							t.Errorf("%s expected error: '%v', got: nil", test.name, lt.expectedErr)
							return
						}

						checkParsedTime(result, expected, test, t)
					})
				}
			}
		})
	}
}

func checkParsedTime(value time.Time, expected time.Time, test ParseTest, t *testing.T) {
	if test.yearSign >= 0 && test.yearSign*value.Year() != expected.Year() {
		t.Errorf("%s: bad year: %d not %d", test.name, value.Year(), expected.Year())
	}
	if value.Month() != expected.Month() {
		t.Errorf("%s: bad month: %s not %s", test.name, value.Month(), expected.Month())
	}
	if value.Day() != expected.Day() {
		t.Errorf("%s: bad day: %d not %d", test.name, value.Day(), expected.Day())
	}
	if value.Hour() != expected.Hour() {
		t.Errorf("%s: bad hour: %d not %d", test.name, value.Hour(), expected.Hour())
	}
	if value.Minute() != expected.Minute() {
		t.Errorf("%s: bad minute: %d not %d", test.name, value.Minute(), expected.Minute())
	}
	if value.Second() != expected.Second() {
		t.Errorf("%s: bad second: %d not %d", test.name, value.Second(), expected.Second())
	}
	if value.Nanosecond() != expected.Nanosecond() {
		t.Errorf("%s: bad nanosecond: %d not %d", test.name, value.Nanosecond(), expected.Nanosecond())
	}

	zoneName, zoneOffset := value.Zone()
	expZoneName, expZoneOffset := expected.Zone()
	if test.hasTZ && zoneOffset != expZoneOffset {
		t.Errorf("%s: bad tz offset: %s %d not %d", test.name, zoneName, zoneOffset, expZoneOffset)
	}
	if test.hasTZ && zoneName != expZoneName {
		t.Errorf("%s: bad tz name: %s not %s", test.name, zoneName, expZoneName)
	}

	if test.hasWD && value.Weekday() != expected.Weekday() {
		t.Errorf("%s: bad weekday: %s not %s", test.name, value.Weekday(), expected.Weekday())
	}
}

func TestLayoutMismatch(t *testing.T) {
	format := "Mon January 03:04:05PM"
	value := "February Mon 03:04:05PM"
	expected := newLayoutMismatchError("Mon", value)

	t.Run("Parse", func(t *testing.T) {
		_, err := Parse(format, value, LocaleEn)
		if err == nil || !errors.Is(err, expected) {
			t.Errorf("ErrLayoutMismatch expected, got %v", err)
		}
	})

	t.Run("ParseInLocation", func(t *testing.T) {
		_, err := ParseInLocation(format, value, LocaleEn, defaultLocation)
		if err == nil || !errors.Is(err, expected) {
			t.Errorf("ErrLayoutMismatch expected, got %v", err)
		}
	})
}

func TestParsingWithUnsupportedLocale(t *testing.T) {
	lang := "ann"
	format := "Mon January 03:04:05PM"
	value := "Feb Monday 03:04:05PM"

	t.Run("Parse", func(t *testing.T) {
		_, err := Parse(format, value, lang)
		var e *ErrUnsupportedLocale
		if !errors.As(err, &e) {
			t.Errorf("expected error: '%v', got: '%v'", &ErrUnsupportedLocale{lang}, err)
		}
	})

	t.Run("ParseInLocation", func(t *testing.T) {
		_, err := ParseInLocation(format, value, lang, defaultLocation)
		var e *ErrUnsupportedLocale
		if !errors.As(err, &e) {
			t.Errorf("expected error: '%v', got: '%v'", &ErrUnsupportedLocale{lang}, err)
		}
	})
}

func TestDayPeriodsLayoutCase(t *testing.T) {
	tests := []struct {
		name   string
		format string
		value  string
		lang   string
	}{
		{
			name:   "AllLowerPm",
			format: "Monday January 03:04:05pm",
			value:  "lunes enero 03:04:05a.m.",
			lang:   LocaleEs,
		},
		{
			name:   "AllUpperPm",
			format: "Monday January 03:04:05PM",
			value:  "Monday January 03:04:05AM",
			lang:   LocaleEnUS,
		},
		{
			name:   "UpperPmLowerValue",
			format: "Monday January 03:04:05PM",
			value:  "Monday January 03:04:05am",
			lang:   LocaleEnUS,
		},
		{
			name:   "LowerPmUpperValue",
			format: "Monday January 03:04:05pm",
			value:  "Monday January 03:04:05AM",
			lang:   LocaleEnUS,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("ParseWith%s", test.name), func(t *testing.T) {
			_, err := Parse(test.format, test.value, test.lang)
			if err != nil {
				t.Errorf("no error expected, got: '%v'", err)
			}
		})

		t.Run(fmt.Sprintf("ParseInLocationWith%s", test.name), func(t *testing.T) {
			_, err := ParseInLocation(test.format, test.value, test.lang, defaultLocation)
			if err != nil {
				t.Errorf("no error expected, got: '%v'", err)
			}
		})
	}
}

func TestAllLocalesReplacements(t *testing.T) {
	var shortLayoutTests []ParseTest
	var longLayoutTests []ParseTest

	for month := 0; month < 12; month++ {
		day := month % 7
		period := month % 2

		shortStdValue := fmt.Sprintf("%s %s 03:04:05%s", shortDayNamesStd[day], shortMonthNamesStd[month], dayPeriodsStd[period])
		shortLayoutTests = append(shortLayoutTests, ParseTest{
			name:     shortStdValue,
			format:   "Mon Jan 03:04:05PM",
			stdValue: shortStdValue,
			hasWD:    true,
			yearSign: -1,
			locales:  allLocalesTests("%s %s 03:04:05%s", []replacement{{shortDayNamesField, day}, {shortMonthNamesField, month}, {dayPeriodsField, period}}),
		})

		longStdValue := fmt.Sprintf("%s %s 03:04:05%s", longDayNamesStd[day], longMonthNamesStd[month], dayPeriodsStd[period])
		longLayoutTests = append(longLayoutTests, ParseTest{
			format:   "Monday January 03:04:05PM",
			stdValue: longStdValue,
			hasWD:    true,
			yearSign: -1,
			locales:  allLocalesTests("%s %s 03:04:05%s", []replacement{{longDayNamesField, day}, {longMonthNamesField, month}, {dayPeriodsField, period}}),
		})
	}

	t.Run("ParseShortNames", func(t *testing.T) {
		testParseFunc(t, shortLayoutTests, time.Parse, Parse)
	})

	t.Run("ParseLongNames", func(t *testing.T) {
		testParseFunc(t, shortLayoutTests, time.Parse, Parse)
	})

	t.Run("ParseInLocationShortNames", func(t *testing.T) {
		testParseFunc(t, longLayoutTests, func(format, value string) (time.Time, error) {
			return time.ParseInLocation(format, value, defaultLocation)
		}, func(format string, value string, locale string) (time.Time, error) {
			return ParseInLocation(format, value, locale, defaultLocation)
		})
	})

	t.Run("ParseInLocationLongNames", func(t *testing.T) {
		testParseFunc(t, longLayoutTests, func(format, value string) (time.Time, error) {
			return time.ParseInLocation(format, value, defaultLocation)
		}, func(format string, value string, locale string) (time.Time, error) {
			return ParseInLocation(format, value, locale, defaultLocation)
		})
	})
}

func TestUnsupportedLayoutElements(t *testing.T) {
	locale := genericLocale{
		lang: LocaleEn,
		table: [5][]string{
			{}, {}, {}, {}, {},
		},
	}

	doTest := func(t *testing.T, layout, elem string) {
		expectedErr := newUnsupportedLayoutElemError(elem, &locale)
		_, err := ParseWithLocale(layout, layout, &locale)
		if err == nil {
			t.Errorf("expected error: '%v', got: nil", expectedErr)
			return
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error: '%v', got: '%v'", expectedErr, err)
		}
	}

	t.Run("ShortWeekDayName", func(t *testing.T) {
		doTest(t, "Mon February 2006", "Mon")
	})

	t.Run("LongWeekDayName", func(t *testing.T) {
		doTest(t, "Monday February 2006", "Monday")
	})

	t.Run("ShortMonthName", func(t *testing.T) {
		doTest(t, "Jan 2006", "Jan")
	})

	t.Run("LongMonthName", func(t *testing.T) {
		doTest(t, "January 2006", "January")
	})

	t.Run("DayPeriod", func(t *testing.T) {
		doTest(t, "03:04:05PM February 2006 ", "PM")
	})
}

func TestTranslate(t *testing.T) {
	translate, err := Translate("Monday _2 2006 27", "viernes 15 2006 27", LocaleEsES)
	if err != nil {
		return
	}

	if translate != "Friday 15 2006 27" {
		t.Errorf("expected value \"Friday 15 2006 27\", got: \"%s\"", translate)
	}
}

func TestTranslateSpacedValue(t *testing.T) {
	locale, err := NewDefaultLocale(LocaleEsES)
	if err != nil {
		return
	}

	translate, err := TranslateWithLocale("Monday _2 2006 27", "    viernes   15  2006 27 ", locale)
	if err != nil {
		return
	}

	if translate != "    Friday   15  2006 27 " {
		t.Errorf("expected spced value '    Friday   15  2006 27 ', got: '%s'", translate)
	}
}

func TestTranslateWithUnsupportedLocale(t *testing.T) {
	lang := "ann"
	format := "Mon January 03:04:05PM"
	value := "Feb Monday 03:04:05PM"

	_, err := Translate(format, value, lang)
	var e *ErrUnsupportedLocale
	if !errors.As(err, &e) {
		t.Errorf("expected error: '%v', got: '%v'", &ErrUnsupportedLocale{lang}, err)
	}
}
