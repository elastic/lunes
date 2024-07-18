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

package benchmarks_test

import (
	"github.com/elastic/lunes"
	"github.com/goodsign/monday"
	"golang.org/x/text/language"
	"testing"
	"time"
)

func BenchmarkTranslate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	locale, err := lunes.NewDefaultLocale(&language.Spanish)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = lunes.Translate("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", locale)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	locale, err := lunes.NewDefaultLocale(&language.EuropeanSpanish)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = lunes.Parse("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", locale)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParseInLocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	locale, err := lunes.NewDefaultLocale(&language.EuropeanSpanish)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, err = lunes.ParseInLocation("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", locale, time.UTC)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParseMonday(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := monday.Parse("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", monday.LocaleEsES)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParseInLocationMonday(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := monday.ParseInLocation("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", time.UTC, monday.LocaleEsES)
		if err != nil {
			b.Error(err)
		}
	}
}
