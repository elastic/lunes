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
	"strings"
	"time"
	"unicode"
)

var longDayNamesStd = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var shortDayNamesStd = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var shortMonthNamesStd = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

var longMonthNamesStd = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

var dayPeriodsStd = []string{
	"AM",
	"PM",
}

// Parse translates the localized time value using the Translate method, and parses the
// output to [time.Time] by using the [time.Parse] method.
func Parse(layout string, value string, locale Locale) (time.Time, error) {
	pv, err := Translate(layout, value, locale)
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(layout, pv)
}

// ParseInLocation translates the localized time value using the Translate method, and parses
// the output to [time.Time] by using the [time.ParseInLocation] method.
func ParseInLocation(layout string, value string, locale Locale, location *time.Location) (time.Time, error) {
	pv, err := Translate(layout, value, locale)
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation(layout, pv, location)
}

// Translate parses a localized textual time value from the provided locale to English.
// It replaces short and long week days names, months names, and day periods by their
// equivalents. The first argument must be a native Go time layout. The second argument
// must be parseable using the format string (layout) provided as the first argument,
// but in the foreign language.
//
// This function is meant to return a value that can be used with the Go standard
// [time.Parse] or [time.ParseInLocation] methods. Although it maintains value's empty
// spaces that are not present in the layout string, it might drop them in the future,
// as they are ignored by both standard time parsings functions.
func Translate(layout string, value string, locale Locale) (string, error) {
	var err error
	var sb strings.Builder
	var layoutOffset, valueOffset int

	sb.Grow(len(layout) + 32)

	for layoutOffset < len(layout) {
		written := false
		var lookupTab, stdTab []string

		switch c := int(layout[layoutOffset]); c {
		case 'J': // January, Jan
			if len(layout) >= layoutOffset+3 && layout[layoutOffset:layoutOffset+3] == "Jan" {
				layoutElem := ""
				if len(layout) >= layoutOffset+7 && layout[layoutOffset:layoutOffset+7] == "January" {
					layoutElem = "January"
					lookupTab = locale.LongMonthNames()
					stdTab = longMonthNamesStd
				} else if !startsWithLowerCase(layout[layoutOffset+3:]) {
					layoutElem = "Jan"
					lookupTab = locale.ShortMonthNames()
					stdTab = shortMonthNamesStd
				}

				if layoutElem == "" {
					break
				}

				if len(lookupTab) == 0 {
					return "", newUnsupportedLayoutElemError(layoutElem, locale)
				}

				layoutOffset += len(layoutElem)
				valueOffset, err = writeLayoutValue(layoutElem, lookupTab, stdTab, valueOffset, value, &sb)
				if err != nil {
					return "", err
				}

				written = true
			}
		case 'M': // Monday, Mon
			if len(layout) >= layoutOffset+3 && layout[layoutOffset:layoutOffset+3] == "Mon" {
				layoutElem := ""
				if len(layout) >= layoutOffset+6 && layout[layoutOffset:layoutOffset+6] == "Monday" {
					layoutElem = "Monday"
					lookupTab = locale.LongDayNames()
					stdTab = longDayNamesStd
				} else if !startsWithLowerCase(layout[layoutOffset+3:]) {
					layoutElem = "Mon"
					lookupTab = locale.ShortDayNames()
					stdTab = shortDayNamesStd
				}

				if layoutElem == "" {
					break
				}

				if len(lookupTab) == 0 {
					return "", newUnsupportedLayoutElemError(layoutElem, locale)
				}

				layoutOffset += len(layoutElem)
				valueOffset, err = writeLayoutValue(layoutElem, lookupTab, stdTab, valueOffset, value, &sb)
				if err != nil {
					return "", err
				}
				written = true
			}
		case 'P', 'p': // PM, pm
			if len(layout) >= layoutOffset+2 && unicode.ToUpper(rune(layout[layoutOffset+1])) == 'M' {
				lookupTab = locale.DayPeriods()
				if len(lookupTab) == 0 {
					return "", newUnsupportedLayoutElemError("PM", locale)
				}

				layoutOffset += 2
				valueOffset, err = writeLayoutValue("PM", lookupTab, dayPeriodsStd, valueOffset, value, &sb)
				if err != nil {
					return "", err
				}
				written = true
			}
		case '_': // _2, _2006, __2
			// Although no translations happens here, it is still necessary to calculate the
			// variable size of `_`  values, so the layoutOffset stays synchronized with
			// its layout element counterpart.
			if len(layout) >= layoutOffset+2 && layout[layoutOffset+1] == '2' {
				var layoutElemSize int
				// _2006 is really a literal _, followed by the long year placeholder
				if len(layout) >= layoutOffset+5 && layout[layoutOffset+1:layoutOffset+5] == "2006" {
					if len(value) >= valueOffset+5 {
						layoutElemSize = 5 // _2006
					}
				} else {
					if len(value) >= valueOffset+2 {
						layoutElemSize = 2 // _2
					}
				}

				if layoutElemSize > 0 {
					layoutOffset += layoutElemSize
					valueOffset, err = writeNextNonSpaceValue(value, valueOffset, layoutElemSize, &sb)
					if err != nil {
						return "", err
					}
					written = true
				}
			}

			if len(layout) >= layoutOffset+3 && layout[layoutOffset+1] == '_' && layout[layoutOffset+2] == '2' {
				if len(value) >= valueOffset+3 {
					layoutOffset += 3
					valueOffset, err = writeNextNonSpaceValue(value, valueOffset, 3, &sb)
					if err != nil {
						return "", err
					}
					written = true
				}
			}
		}

		if !written {
			var writtenSize int
			if len(value) > valueOffset {
				writtenSize, err = sb.WriteRune(rune(value[valueOffset]))
				if err != nil {
					return "", err
				}
			}

			layoutOffset++
			valueOffset += writtenSize
		}
	}

	if len(value) >= valueOffset {
		sb.WriteString(value[valueOffset:])
	}

	return sb.String(), nil
}

func writeNextNonSpaceValue(value string, offset int, max int, sb *strings.Builder) (int, error) {
	nextValOffset, skippedSpaces, val, err := nextNonSpaceValue(value, offset, max)
	if err != nil {
		return offset, err
	}

	if skippedSpaces > 0 {
		val = strings.Repeat(" ", skippedSpaces) + val
	}

	_, err = sb.WriteString(val)
	if err != nil {
		return offset, err
	}

	return nextValOffset, nil
}

func writeLayoutValue(layoutElem string, lookupTab, stdTab []string, valueOffset int, value string, sb *strings.Builder) (int, error) {
	newOffset, skippedSpaces, foundStdValue, val := lookup(lookupTab, valueOffset, value, stdTab)
	if foundStdValue == "" {
		return valueOffset, newLayoutMismatchError(layoutElem, value)
	}

	if skippedSpaces > 0 {
		foundStdValue = strings.Repeat(" ", skippedSpaces) + foundStdValue
	}

	_, err := sb.WriteString(foundStdValue)
	if err != nil {
		return valueOffset, err
	}

	newOffset += len(val)
	return newOffset, nil
}

func nextNonSpaceValue(value string, offset int, max int) (newOffset, skippedSpaces int, val string, err error) {
	newOffset = offset
	for newOffset < len(value) && unicode.IsSpace(rune(value[newOffset])) {
		newOffset++
	}

	skippedSpaces = newOffset - offset
	if newOffset > len(value) {
		return offset, skippedSpaces, "", errors.New("next non-space value not found")
	}

	for newOffset < len(value) {
		if !unicode.IsSpace(rune(value[newOffset])) {
			val += string(value[newOffset])
			newOffset++
		} else {
			return newOffset, skippedSpaces, val, nil
		}

		if len(val) == max {
			return newOffset, skippedSpaces, val, nil
		}
	}

	return newOffset, skippedSpaces, val, nil
}

func lookup(lookupTab []string, offset int, val string, stdTab []string) (newOffset, skippedSpaces int, stdValue string, value string) {
	newOffset = offset
	for newOffset < len(val) && unicode.IsSpace(rune(val[newOffset])) {
		newOffset++
	}

	skippedSpaces = newOffset - offset
	if newOffset > len(val) {
		return offset, skippedSpaces, "", val
	}

	for i, v := range lookupTab {
		// Already matched a more specific/longer value
		if stdValue != "" && len(v) <= len(value) {
			continue
		}

		end := newOffset + len(v)
		if end > len(val) {
			continue
		}

		candidate := val[newOffset:end]
		if len(candidate) == len(v) && strings.EqualFold(candidate, v) {
			stdValue = stdTab[i]
			value = candidate
		}
	}

	return newOffset, skippedSpaces, stdValue, value
}

func startsWithLowerCase(value string) bool {
	if len(value) == 0 {
		return false
	}
	c := value[0]
	return 'a' <= c && c <= 'z'
}

// ErrLayoutMismatch indicates that a provided value does not match its layout counterpart.
type ErrLayoutMismatch struct {
	Value      string
	LayoutElem string
}

func (l *ErrLayoutMismatch) Error() string {
	return fmt.Sprintf(`value "%s" does not match the layout element "%s"`, l.Value, l.LayoutElem)
}

func (l *ErrLayoutMismatch) Is(err error) bool {
	var target *ErrLayoutMismatch
	if ok := errors.As(err, &target); ok {
		return l.Value == target.Value && l.LayoutElem == target.LayoutElem
	}
	return false
}

func newLayoutMismatchError(elem, value string) error {
	return &ErrLayoutMismatch{
		LayoutElem: elem,
		Value:      value,
	}
}

// ErrUnsupportedLayoutElem indicates that a provided layout element is not supported by
// the given locale/language.
type ErrUnsupportedLayoutElem struct {
	LayoutElem string
	Language   string
}

func (u *ErrUnsupportedLayoutElem) Error() string {
	return fmt.Sprintf(`layout element "%s" is not support by the language "%s"`, u.LayoutElem, u.Language)
}

func (u *ErrUnsupportedLayoutElem) Is(err error) bool {
	var target *ErrUnsupportedLayoutElem
	if ok := errors.As(err, &target); ok {
		return u.Language == target.Language && u.LayoutElem == target.LayoutElem
	}
	return false
}

func newUnsupportedLayoutElemError(elem string, locale Locale) error {
	return &ErrUnsupportedLayoutElem{
		LayoutElem: elem,
		Language:   locale.Language().String(),
	}
}
