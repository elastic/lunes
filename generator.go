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

//go:build ignore

package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/text/language"
)

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

var shortDayNamesStd = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var longDayNamesStd = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var dayPeriodsStd = []string{
	"AM",
	"PM",
}

var longDayNamesStdMap = map[string]string{
	"sun": "Sunday",
	"mon": "Monday",
	"tue": "Tuesday",
	"wed": "Wednesday",
	"thu": "Thursday",
	"fri": "Friday",
	"sat": "Saturday",
}

var shortDayNamesStdMap = map[string]string{
	"sun": "Sun",
	"mon": "Mon",
	"tue": "Tue",
	"wed": "Wed",
	"thu": "Thu",
	"fri": "Fri",
	"sat": "Sat",
}

var dayPeriodsStdMap = map[string]string{
	"am": "AM",
	"pm": "PM",
}

var localesData = map[string]*cldrLocaleData{}

func main() {
	cldrVersion := flag.Int("cldr", 45, "CLDR version")
	cldrZipFilePath := flag.String("file", "", "CLDR core.zip path")
	flag.Parse()

	data, err := readCLDRCoreFile(*cldrZipFilePath, *cldrVersion)
	if err != nil {
		log.Fatalf("failed to read CLDR zip: %v", err)
	}

	sortedLanguages := buildLanguageGraph(data).getSorted()
	var nonEmptyLanguages []string
	for _, tag := range sortedLanguages {
		localeLDML := data[tag]
		parsedTag := language.Make(tag)

		var localeCalendar cldrLocaleData
		if parsedTag.Parent() != language.Und {
			existing, ok := localesData[localeLDML.Parent]
			if ok {
				localeCalendar = existing.clone()
			} else {
				localeCalendar = cldrLocaleData{}
			}
		} else {
			localeCalendar = cldrLocaleData{}
		}

		gregorianCalendar := findGregorianCalendar(localeLDML.LDML)
		if gregorianCalendar != nil {
			err = fillLocaleData(tag, gregorianCalendar, &localeCalendar)
			if err != nil {
				log.Fatal(err)
			}
		}

		if !localeCalendar.isEmpty() {
			localesData[tag] = &localeCalendar
			nonEmptyLanguages = append(nonEmptyLanguages, tag)
		} else {
			log.Printf("skipped locale with empty dates: %s\n", tag)
		}
	}

	sort.Strings(nonEmptyLanguages)

	var tablesTmplDataItems []*tablesTmplDataItem
	for _, tag := range nonEmptyLanguages {
		localeCalendar := localesData[tag]
		if localeCalendar.isValid() {
			tablesTmplDataItems = append(tablesTmplDataItems, newTablesTmplDataItem(tag, localeCalendar))
		} else {
			log.Printf("skipped invalid locale: %s\n", tag)
		}
	}

	err = writeTableGoFile(cldrVersion, tablesTmplDataItems)
	if err != nil {
		log.Fatal(err)
	}
}

func fillLocaleData(tag string, gregorianCalendar *Calendar, locale *cldrLocaleData) error {
	var err error
	if gregorianCalendar.Months != nil && gregorianCalendar.Months.MonthContext != nil {
		for _, monthContext := range gregorianCalendar.Months.MonthContext {
			if monthContext.Type != "format" {
				continue
			}

			for _, monthWidth := range monthContext.MonthWidth {
				if monthWidth.Type == "abbreviated" {
					locale.shortMonthNames, err = lookupMonthValue(locale.shortMonthNames, shortMonthNamesStd, monthWidth.Month)
					if err != nil {
						return fmt.Errorf("failed to read %s short month names %w", tag, err)
					}
				} else if monthWidth.Type == "wide" {
					locale.longMonthNames, err = lookupMonthValue(locale.longMonthNames, longMonthNamesStd, monthWidth.Month)
					if err != nil {
						return fmt.Errorf("failed to read %s long month names %w", tag, err)
					}
				}
			}
		}
	}

	if gregorianCalendar.Days != nil && gregorianCalendar.Days.DayContext != nil {
		for _, dayContext := range gregorianCalendar.Days.DayContext {
			if dayContext.Type != "format" {
				continue
			}

			for _, dayWidth := range dayContext.DayWidth {
				if dayWidth.Type == "abbreviated" {
					locale.shortDayNames = lookupDayValue(locale.shortDayNames, shortDayNamesStdMap, dayWidth.Day)
				} else if dayWidth.Type == "wide" {
					locale.longDayNames = lookupDayValue(locale.longDayNames, longDayNamesStdMap, dayWidth.Day)
				}
			}
		}
	}

	if gregorianCalendar.DayPeriods != nil && gregorianCalendar.DayPeriods.DayPeriodContext != nil {
		for _, periodContext := range gregorianCalendar.DayPeriods.DayPeriodContext {
			if periodContext.Type != "format" {
				continue
			}

			periods := map[string]string{}
			for _, periodWidth := range periodContext.DayPeriodWidth {
				if periodWidth.Type != "abbreviated" && periodWidth.Type != "narrow" {
					continue
				}

				for _, period := range periodWidth.DayPeriod {
					if p, ok := dayPeriodsStdMap[period.Type]; ok {
						// preference for non-variant periods
						if _, ok = periods[p]; ok && period.Alt == "variant" {
							continue
						}

						periods[p] = strings.ReplaceAll(period.CharData, "\u202F", "")
					}
				}

				if len(periods) == 2 {
					locale.dayPeriods = periods
					break
				}
			}
		}
	}

	return nil
}

func lookupMonthValue(curr map[string]string, stdTab []string, lookupTable []*MonthWidth) (map[string]string, error) {
	if curr == nil && len(lookupTable) == 0 {
		return nil, nil
	}

	val := make(map[string]string, 12)
	if curr != nil {
		maps.Copy(val, curr)
	}

	for _, month := range lookupTable {
		m, err := strconv.Atoi(month.Type)
		if err != nil {
			return nil, err
		}

		val[stdTab[m-1]] = month.CharData
	}

	return val, nil
}

func lookupDayValue(curr map[string]string, stdTab map[string]string, lookupTable []*Common) map[string]string {
	if curr == nil && len(lookupTable) == 0 {
		return nil
	}

	val := make(map[string]string, 7)
	if curr != nil {
		maps.Copy(val, curr)
	}

	for _, day := range lookupTable {
		d, ok := stdTab[day.Type]
		if !ok {
			continue
		}

		val[d] = day.CharData
	}

	return val
}

func findGregorianCalendar(lang *LDML) *Calendar {
	if lang == nil || lang.Dates == nil || lang.Dates.Calendars == nil || lang.Dates.Calendars.Calendar == nil {
		return nil
	}

	for _, calendar := range lang.Dates.Calendars.Calendar {
		if calendar.Type == "gregorian" {
			return calendar
		}
	}

	return nil
}

func readCLDRCoreFile(path string, version int) (map[string]*cldrLocaleModel, error) {
	cldrCoreZipFile, err := getCLDRCoreFile(path, version)
	if err != nil {
		return nil, err
	}

	defer cldrCoreZipFile.Close()

	zipFile, err := zip.OpenReader(cldrCoreZipFile.Name())
	if err != nil {
		return nil, err
	}

	defer zipFile.Close()

	models := make(map[string]*cldrLocaleModel)
	for _, file := range zipFile.File {
		fileInfo := file.FileInfo()
		if strings.HasPrefix(file.Name, "common/main") && !fileInfo.IsDir() {
			if strings.HasSuffix(fileInfo.Name(), ".xml") {
				model := &LDML{}
				entry, err := file.Open()
				if err != nil {
					return nil, err
				}

				decoder := xml.NewDecoder(entry)
				if err = decoder.Decode(model); err != nil {
					return nil, err
				}

				tag := fileInfo.Name()[:len(fileInfo.Name())-4]
				parsedTag, err := language.Parse(tag)
				if err != nil {
					return nil, err
				}

				var parent string
				if parsedTag.Parent() != language.Und {
					parent = parsedTag.Parent().String()
				}

				models[parsedTag.String()] = &cldrLocaleModel{parent, model}
			}
		}
	}

	return models, nil
}

func getCLDRCoreFile(path string, version int) (*os.File, error) {
	var cldrCoreZipFile *os.File
	var err error
	if path != "" {
		cldrCoreZipFile, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open CLDR file: %w", err)
		}
		return cldrCoreZipFile, nil
	}

	cldrCoreZipFile, err = downloadCLDRCoreFile(version)
	if err != nil {
		return nil, fmt.Errorf("failed to download CLDR file: %w", err)
	}

	return os.Open(cldrCoreZipFile.Name())
}

func downloadCLDRCoreFile(version int) (file *os.File, err error) {
	tmpFile, err := os.CreateTemp("", "cldr-core*.zip")
	if err != nil {
		return nil, err
	}

	defer tmpFile.Close()

	url := fmt.Sprintf("https://unicode.org/Public/cldr/%d/core.zip", version)
	resp, err := http.Get(url) //nolint:gosec,noctx //not unsafe
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CLDR download failed with status %s", resp.Status)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return nil, err
	}

	return tmpFile, nil
}

type cldrLocaleData struct {
	longDayNames    map[string]string
	shortDayNames   map[string]string
	longMonthNames  map[string]string
	shortMonthNames map[string]string
	dayPeriods      map[string]string
}

func (g *cldrLocaleData) clone() cldrLocaleData {
	return cldrLocaleData{
		longDayNames:    maps.Clone(g.longDayNames),
		shortDayNames:   maps.Clone(g.shortDayNames),
		longMonthNames:  maps.Clone(g.longMonthNames),
		shortMonthNames: maps.Clone(g.shortMonthNames),
		dayPeriods:      maps.Clone(g.dayPeriods),
	}
}

func (g *cldrLocaleData) isEmpty() bool {
	return g.shortDayNames == nil &&
		g.longDayNames == nil &&
		g.shortMonthNames == nil &&
		g.longMonthNames == nil &&
		g.dayPeriods == nil
}

func (g *cldrLocaleData) isValid() bool {
	if g.isEmpty() {
		return false
	}

	if g.shortDayNames != nil && len(g.shortDayNames) != 7 {
		return false
	}

	if g.longDayNames != nil && len(g.longDayNames) != 7 {
		return false
	}

	if g.shortMonthNames != nil && len(g.shortMonthNames) != 12 {
		return false
	}

	if g.longMonthNames != nil && len(g.longMonthNames) != 12 {
		return false
	}

	if g.dayPeriods != nil && len(g.dayPeriods) != 2 {
		return false
	}

	return true
}

type tablesTmplData struct {
	CLDRVersion *int
	Tables      []*tablesTmplDataItem
}

type tablesTmplDataItem struct {
	Name            string
	Language        string
	ShortDaysNames  []string
	LongDaysNames   []string
	ShortMonthNames []string
	LongMonthNames  []string
	DayPeriods      []string
}

func newTablesTmplDataItem(tag string, data *cldrLocaleData) *tablesTmplDataItem {
	name := strings.ReplaceAll(tag, "-", "")
	name = strings.ToUpper(name[:1]) + name[1:]

	return &tablesTmplDataItem{
		Name:            name,
		Language:        tag,
		ShortDaysNames:  sortTableValues(data.shortDayNames, shortDayNamesStd),
		LongDaysNames:   sortTableValues(data.longDayNames, longDayNamesStd),
		ShortMonthNames: sortTableValues(data.shortMonthNames, shortMonthNamesStd),
		LongMonthNames:  sortTableValues(data.longMonthNames, longMonthNamesStd),
		DayPeriods:      sortTableValues(data.dayPeriods, dayPeriodsStd),
	}
}

func sortTableValues(table map[string]string, keys []string) []string {
	if table == nil {
		return []string{}
	}

	sorted := make([]string, 0, len(keys))
	for _, m := range keys {
		sorted = append(sorted, table[m])
	}

	return sorted
}

func stringSliceValue(values []string) string {
	if values == nil {
		return ""
	}

	var sb strings.Builder
	for i, v := range values {
		sb.WriteString(fmt.Sprintf(`"%s"`, v))
		if i+1 < len(values) {
			sb.WriteString(", ")
		}
	}

	return sb.String()
}

func writeTableGoFile(cldrVersion *int, tables []*tablesTmplDataItem) error {
	data := tablesTmplData{
		CLDRVersion: cldrVersion,
		Tables:      tables,
	}

	tablesTmpl := filepath.Join("templates", "tables.go.tmpl")
	tmpl, err := template.New(path.Base(tablesTmpl)).
		Funcs(template.FuncMap{"StringSliceValue": stringSliceValue}).
		ParseFiles(tablesTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse tables.go template: %w", err)
	}

	tablesOutPath := filepath.Join("tables.go")
	tablesOutFile, err := os.OpenFile(tablesOutPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", tablesOutPath, err)
	}

	defer tablesOutFile.Close()

	tmplBuffer := bytes.Buffer{}
	err = tmpl.Execute(&tmplBuffer, data)
	if err != nil {
		return fmt.Errorf("failed to execute tables.go template: %w", err)
	}

	formattedBuffer, err := format.Source(tmplBuffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format tables.go template output: %w", err)
	}

	_, err = tablesOutFile.Write(formattedBuffer)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", tablesOutPath, err)
	}

	return nil
}

type cldrLocaleModel struct {
	Parent string
	*LDML
}

// cldrGraph is used to sort the languages tags considering its upper level dependencies,
// ensuring that base languages are parsed first, so derivatives can copy the data.
// E.g.: ["en", "en-001", "en-AU"]
type cldrGraph struct {
	vertices []string
	edges    map[string][]*cldrGraphEdge
}

type cldrGraphEdge struct {
	from string
	to   string
}

func (c *cldrGraph) add(lang, parent string) {
	c.vertices = append(c.vertices, lang)
	if parent != "" {
		c.edges[lang] = append(c.edges[lang], &cldrGraphEdge{lang, parent})
	}
}

func (c *cldrGraph) getSorted() []string {
	visited := map[string]bool{}
	var stack []string

	for _, node := range c.vertices {
		if _, ok := visited[node]; !ok {
			c.dfs(node, visited, &stack)
		}
	}

	return stack
}

func (c *cldrGraph) dfs(from string, visited map[string]bool, stack *[]string) {
	visited[from] = true
	edges, ok := c.edges[from]
	if ok {
		for _, edge := range edges {
			if _, ok = visited[edge.to]; !ok {
				c.dfs(edge.to, visited, stack)
			}
		}
	}

	*stack = append(*stack, from)
}

func buildLanguageGraph(models map[string]*cldrLocaleModel) *cldrGraph {
	graph := &cldrGraph{
		vertices: []string{},
		edges:    make(map[string][]*cldrGraphEdge),
	}

	for tag, model := range models {
		graph.add(tag, model.Parent)
	}

	return graph
}

// Common holds several of the most common attributes and sub elements of an XML element.
type Common struct {
	XMLName         xml.Name
	Type            string `xml:"type,attr,omitempty"`
	Reference       string `xml:"reference,attr,omitempty"`
	Alt             string `xml:"alt,attr,omitempty"`
	ValidSubLocales string `xml:"validSubLocales,attr,omitempty"`
	Draft           string `xml:"draft,attr,omitempty"`
	hidden
}

type hidden struct {
	CharData string `xml:",chardata"`
	Alias    *struct {
		Common
		Source string `xml:"source,attr"`
		Path   string `xml:"path,attr"`
	} `xml:"alias"`
	Def *struct {
		Common
		Choice string `xml:"choice,attr,omitempty"`
		Type   string `xml:"type,attr,omitempty"`
	} `xml:"default"`
}

// LDML is the top-level type for locale-specific data.
type LDML struct {
	Common
	Dates *struct {
		Common
		Calendars *struct {
			Common
			Calendar []*Calendar `xml:"calendar"`
		} `xml:"calendars"`
	} `xml:"dates"`
}

// Calendar specifies the fields used for formatting and parsing dates and times.
type Calendar struct {
	Common
	Months *struct {
		Common
		MonthContext []*struct {
			Common
			MonthWidth []*struct {
				Common
				Month []*MonthWidth `xml:"month"`
			} `xml:"monthWidth"`
		} `xml:"monthContext"`
	} `xml:"months"`
	Days *struct {
		Common
		DayContext []*struct {
			Common
			DayWidth []*struct {
				Common
				Day []*Common `xml:"day"`
			} `xml:"dayWidth"`
		} `xml:"dayContext"`
	} `xml:"days"`
	DayPeriods *struct {
		Common
		DayPeriodContext []*struct {
			Common
			DayPeriodWidth []*struct {
				Common
				DayPeriod []*Common `xml:"dayPeriod"`
			} `xml:"dayPeriodWidth"`
		} `xml:"dayPeriodContext"`
	} `xml:"dayPeriods"`
}

type MonthWidth = struct {
	Common
	Yeartype string `xml:"yeartype,attr"`
}
