# Lunes

---

**Lunes** is a [Go](http://golang.org) library for parsing localized time strings into `time.Time` objects.

There's no intention to replace the standard `time` package parsing functions, instead, it acts as wrapper
translating the provided value to English before invoking the `time.Parse` and `time.ParseInLocation`.

It currently supports almost all [CLDR](https://cldr.unicode.org/) core locales (+900 including drafts),
being limited to the **gregorian** calendars.

Once the official Go i18n features for time parsing are ready, it should be replaced.

## Usage

#### Parse

```go
// creates a new generic locale based on the CLDR gregorian calendar translations, and on
// the provided BCP 47 language.Tag. It results in an ErrUnsupportedLocale if the language.Tag
// is unknown and/or no default data is found.
locale, err := lunes.NewDefaultLocale(&language.EuropeanSpanish)

// it's like time.Parse, but with an additional param "locale" to perform the value translation.
// If the given locale does not support any layout element specified on the layout argument,
// it results in an ErrUnsupportedLayoutElem error. On the other hand, if the values does
// not match the layout, an ErrLayoutMismatch is returned.
t, err := lunes.Parse("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", locale)

// parse in specific time zones.
t, err := lunes.ParseInLocation("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", locale, time.UTC)
```

#### Translate

```go
// creates a new generic locale based on the CLDR gregorian calendar translations.
locale, err := lunes.NewDefaultLocale(&language.EuropeanSpanish)

// translates the value, without parsing to time.Time. If the given locale does not support
// any layout element specified on the layout argument, it results in an ErrUnsupportedLayoutElem
// error. On the other hand, if the values does not match the layout, an ErrLayoutMismatch is returned.
// For the following example, it results in: Friday Jan 27 11:53:29.
str, err := lunes.Translate("Monday Jan _2 15:04:05", "viernes ene 27 11:53:29", locale)

// the translated value is meant to be used with the time package functions
t, err := time.Parse("Monday Jan _2 15:04:05", str)
```

#### Custom locales

A `lunes.Locale` provides a collection of time layouts translations to a specific language.
It is used to provide a map between those translations and the English language.
In oder to use custom locales, the following functions must be implemented:

```go
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
```

Then, the custom locale can be passes as argument to the `lunes.ParseInLocation` function:

```go
customLocale := &CustomLocale{}

t, err := lunes.ParseInLocation("Monday Jan _2 2006 15:04:05", "lunes oct 27 1988 11:53:29", myLocale)
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

