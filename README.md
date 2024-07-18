# Lunes

---

**Lunes** is a [Go](http://golang.org) library for parsing localized time strings into `time.Time` objects.

There's no intention to replace the standard `time` package parsing functions, instead, it acts as wrapper
translating the provide value to English before invoking the `time.Parse` and `time.ParseInLocation`.

It currently supports almost all [CLDR](https://cldr.unicode.org/) core locales (+900 including drafts),
being limited to the **gregorian** calendars.

Once the official Go i18n features for time parsing are ready, it should be replaced.

## Usage

#### Parse

```go
// uses the default CLDR generated data locales
locale, err := lunes.NewDefaultLocale(&language.EuropeanSpanish)

// it's like time.Parse, but with an additional param "locale" to perform the value translation
t, err := lunes.Parse("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", locale)

// parse in specif time zones
t, err := lunes.Parse("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", locale, time.UTC)
```

#### Translate

```go
// uses the default CLDR generated data locales
locale, err := lunes.NewDefaultLocale(&language.EuropeanSpanish)

// translates the value, without parsing to time.Time
// it results in: Friday Jan 27 11:53:29
str, err := lunes.Translate("Monday Jan _2 15:04:05", "viernes ene 27 11:53:29", locale)

// the translated value is meant to be used with the time package functions
t, err := time.Parse("Monday Jan _2 15:04:05", str)
```

#### Custom locales

Custom locales must implement the `lunes.Locale` interface. Please note that the values
returned by this interface must be sorted as described in each method documentation.

```go
myLocale := &MyCustomLocale{}

t, err := lunes.Parse("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", myLocale)
```

## Benchmarks

Comparing to [github.com/goodsign/monday](https://github.com/goodsign/monday)

```
BenchmarkTranslate-10                	 5074365	       232.5 ns/op	      76 B/op	       4 allocs/op
BenchmarkParse-10                    	 3361557	       366.3 ns/op	      76 B/op	       4 allocs/op
BenchmarkParseInLocation-10          	 3397855	       353.6 ns/op	      76 B/op	       4 allocs/op
BenchmarkParseMonday-10              	  217746	      5559 ns/op	    3753 B/op	     117 allocs/op
BenchmarkParseInLocationMonday-10    	  214958	      5577 ns/op	    3753 B/op	     117 allocs/op
```

### Usage notes

- It currently supports the following layout replacements:
  - Short days names (`Mon`)
  - Long days names (`Monday`)
  - Short month names (`Jan`)
  - Long month names (`January`)
  - Day periods (`PM`)
- Translations are auto-generated, and it might be inconsistent depending on the CLDR locale [stage](https://cldr.unicode.org/index/process).
- A few locales does not support (or are missing) translations for specific layout elements (short/long days/month names or day periods), in that case,
  an error will be reported.

