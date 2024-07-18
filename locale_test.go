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
	"strconv"
	"testing"

	"golang.org/x/text/language"
)

func TestLocaleTableFieldsLookups(t *testing.T) {
	shortDaysNameVal := strconv.Itoa(shortDayNamesField)
	longDayNamesVal := strconv.Itoa(longDayNamesField)
	shortMonthNamesVal := strconv.Itoa(shortMonthNamesField)
	longMonthNamesVal := strconv.Itoa(longMonthNamesField)
	dayPeriodsVal := strconv.Itoa(dayPeriodsField)

	locale := genericLocale{
		lang: &language.English,
		table: [5][]string{
			{shortDaysNameVal},
			{longDayNamesVal},
			{shortMonthNamesVal},
			{longMonthNamesVal},
			{dayPeriodsVal},
		},
	}

	if locale.Language() != &language.English {
		t.Errorf("expected Language: %v, got: %v", language.English.String(), locale.Language().String())
	}

	if locale.ShortDayNames()[0] != shortDaysNameVal {
		t.Errorf("expected: %s, got: %s", locale.ShortDayNames()[0], shortDaysNameVal)
	}

	if locale.LongDayNames()[0] != longDayNamesVal {
		t.Errorf("expected: %s, got: %s", locale.ShortDayNames()[0], longDayNamesVal)
	}

	if locale.ShortMonthNames()[0] != shortMonthNamesVal {
		t.Errorf("expected: %s, got: %s", locale.ShortMonthNames()[0], shortMonthNamesVal)
	}

	if locale.LongMonthNames()[0] != longMonthNamesVal {
		t.Errorf("expected: %s, got: %s", locale.LongMonthNames()[0], longMonthNamesVal)
	}

	if locale.DayPeriods()[0] != dayPeriodsVal {
		t.Errorf("expected: %s, got: %s", locale.DayPeriods()[0], dayPeriodsVal)
	}

}

func TestUnsupportedLocale(t *testing.T) {
	ann := language.Make("ann")
	_, err := NewDefaultLocale(&ann)
	if err == nil {
		t.Error("expected ErrUnsupportedLocale error, got: nil")
	}

	expected := &ErrUnsupportedLocale{&ann}

	var e *ErrUnsupportedLocale
	if !errors.As(err, &e) || e.lang != &ann {
		t.Errorf("expected error: '%v', got: '%v'", expected, err)
	}
}
