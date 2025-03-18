package templates

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Volte6/ansitags"
	"github.com/mattn/go-runewidth"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/util"
)

type AnsiFlag uint8

const (
	AnsiTagsDefault  AnsiFlag = iota // Do not parse tags
	AnsiTagsParse                    // Parse ansi tags before returning contents of template
	AnsiTagsStrip                    // strip out all ansi tags and leave text plain
	AnsiTagsMono                     // Parse ansi tags, but strip out all color information
	AnsiTagsPreParse                 // Parse tags before executing the template
)

type cacheEntry struct {
	tpl           *template.Template
	ansiPreparsed bool
	modified      time.Time
}

func (t *cacheEntry) older(compareTime time.Time) bool {
	return t.modified.Before(compareTime)
}

var (
	cacheLock            sync.Mutex
	templateCache        = make(map[string]cacheEntry)
	forceAnsiFlags       = AnsiTagsParse
	ansiLock             sync.RWMutex
	ansiAliasFileModTime time.Time
)

func Exists(name string) bool {

	var fullPath string = util.FilePath(string(configs.GetFilePathsConfig().FolderDataFiles)+`/templates`, `/`, name+`.template`)
	_, err := os.Stat(fullPath)

	return err == nil
}

// Configure a forced ansi flag setting
func SetAnsiFlag(flag AnsiFlag) {
	forceAnsiFlags = flag
}

func Process(name string, data any, ansiFlags ...AnsiFlag) (string, error) {
	ansiLock.RLock()
	defer ansiLock.RUnlock()

	var parseAnsiTags bool = false
	var preParseAnsiTags bool = false

	var ansitagsParseBehavior []ansitags.ParseBehavior = make([]ansitags.ParseBehavior, 0, 5)

	if forceAnsiFlags != AnsiTagsDefault {
		//	ansiFlags = append(ansiFlags, forceAnsiFlags)
	}

	for _, flag := range ansiFlags {
		switch flag {
		case AnsiTagsStrip:
			ansitagsParseBehavior = append(ansitagsParseBehavior, ansitags.StripTags)
		case AnsiTagsMono:
			ansitagsParseBehavior = append(ansitagsParseBehavior, ansitags.Monochrome)
		case AnsiTagsParse:
			parseAnsiTags = true
		case AnsiTagsPreParse:
			preParseAnsiTags = true
		}
	}

	// All templates must end with .template
	var fullPath string = util.FilePath(string(configs.GetFilePathsConfig().FolderDataFiles), `/`, `templates`, `/`, name+`.template`)

	fInfo, err := os.Stat(fullPath)
	if err != nil {
		//mudlog.Error("could not stat template file", "error", err)
		return "[TEMPLATE READ ERROR]", err
	}

	var cache cacheEntry
	var ok bool

	cacheLock.Lock()
	defer cacheLock.Unlock()

	// check if the template is in the cache
	if cache, ok = templateCache[name]; !ok || cache.older(fInfo.ModTime()) {

		// Get the file contents
		fileContents, err := os.ReadFile(fullPath)
		if err != nil {
			mudlog.Error("could not read template file", "error", err)
			return "[TEMPLATE READ ERROR]", err
		}

		if parseAnsiTags && preParseAnsiTags {
			fileContents = []byte(ansitags.Parse(string(fileContents), ansitagsParseBehavior...))
		}

		// parse the file contents as a template
		tpl, err := template.New(name).Funcs(funcMap).Parse(string(fileContents))
		if err != nil {
			return string(fileContents), err
		}

		// add the template to the cache
		cache = cacheEntry{tpl: tpl, modified: fInfo.ModTime(), ansiPreparsed: preParseAnsiTags}
		templateCache[name] = cache
	}

	// execute the template and store the results into a buffer
	var buf bytes.Buffer
	err = cache.tpl.Execute(&buf, data)
	if err != nil {
		mudlog.Error("could not parse template file", "error", err)
		return "[TEMPLATE ERROR]", err
	}

	// return the final data as a string, parse ansi tags if needed (No need to parse if it was preparsed)
	if parseAnsiTags && !cache.ansiPreparsed {
		return ansitags.Parse(buf.String(), ansitagsParseBehavior...), nil
	}

	return buf.String(), nil
}

func ProcessText(text string, data any, ansiFlags ...AnsiFlag) (string, error) {
	var parseAnsiTags bool = false
	var preParseAnsiTags bool = false

	var ansitagsParseBehavior []ansitags.ParseBehavior = make([]ansitags.ParseBehavior, 0, 5)

	if forceAnsiFlags != AnsiTagsDefault {
		//	ansiFlags = append(ansiFlags, forceAnsiFlags)
	}

	for _, flag := range ansiFlags {
		switch flag {
		case AnsiTagsStrip:
			ansitagsParseBehavior = append(ansitagsParseBehavior, ansitags.StripTags)
		case AnsiTagsMono:
			ansitagsParseBehavior = append(ansitagsParseBehavior, ansitags.Monochrome)
		case AnsiTagsParse:
			parseAnsiTags = true
		case AnsiTagsPreParse:
			preParseAnsiTags = true
		}
	}

	if parseAnsiTags && preParseAnsiTags {
		text = ansitags.Parse(text, ansitagsParseBehavior...)
	}

	// parse the file contents as a template
	tpl, err := template.New("").Funcs(funcMap).Parse(text)
	if err != nil {
		return text, err
	}

	// execute the template and store the results into a buffer
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		mudlog.Error("could not parse template text", "error", err)
		return "[TEMPLATE TEXT ERROR]", err
	}

	// return the final data as a string, parse ansi tags if needed (No need to parse if it was preparsed)
	if parseAnsiTags && !preParseAnsiTags {
		return ansitags.Parse(buf.String(), ansitagsParseBehavior...), nil
	}

	return buf.String(), nil
}

const cellPadding int = 1

type TemplateTable struct {
	Title              string
	Header             []string
	Rows               [][]string
	TrueHeaderCellSize []int
	TrueCellSize       [][]int
	ColumnCount        int
	ColumnWidths       []int
	Formatting         [][]string
	formatRowCount     int
}

func (t TemplateTable) GetHeaderCell(column int) string {

	cellStr := t.Header[column]
	repeatCt := t.ColumnWidths[column] - t.TrueHeaderCellSize[column]
	if repeatCt > 0 {
		cellStr += strings.Repeat(` `, repeatCt)
	}

	return cellStr
}

func (t TemplateTable) GetCell(row int, column int) string {

	cellStr := t.Rows[row][column]
	repeatCt := t.ColumnWidths[column] - t.TrueCellSize[row][column]
	if repeatCt > 0 {
		cellStr += strings.Repeat(` `, repeatCt)
	}

	if t.formatRowCount > 0 {
		cellFormat := t.Formatting[row%t.formatRowCount][column]
		if cellFormat[0:1] == `:` {
			return colorpatterns.ApplyColorPattern(cellStr, cellFormat[1:])
		}
		return fmt.Sprintf(t.Formatting[row%t.formatRowCount][column], cellStr)
	}
	return cellStr
}

func GetTable(title string, headers []string, rows [][]string, formatting ...[]string) TemplateTable {

	var table TemplateTable = TemplateTable{
		Title:              title,
		Header:             headers,
		Rows:               rows,
		TrueHeaderCellSize: []int{},
		TrueCellSize:       [][]int{},
		ColumnCount:        len(headers),
		ColumnWidths:       make([]int, len(headers)),
		Formatting:         formatting,
	}

	hdrColCt := len(headers)
	rowCt := len(rows)
	table.formatRowCount = len(formatting)
	table.TrueHeaderCellSize = make([]int, hdrColCt)
	table.TrueCellSize = make([][]int, rowCt)

	// Get the longest element
	for i := 0; i < hdrColCt; i++ {
		sz := runewidth.StringWidth(headers[i])
		if sz+1 > table.ColumnWidths[i] {
			table.ColumnWidths[i] = sz
		}
		table.TrueHeaderCellSize[i] = sz
	}

	// Get the longest element
	for r := 0; r < rowCt; r++ {
		rowColCt := len(rows[r])
		table.TrueCellSize[r] = make([]int, rowColCt)

		if hdrColCt < rowColCt {
			for i := hdrColCt; i < rowColCt; i++ {
				table.Header = append(table.Header, ``)
			}
			hdrColCt = len(table.Header)
		}

		for c := 0; c < hdrColCt; c++ {
			sz := runewidth.StringWidth(ansitags.Parse(rows[r][c], ansitags.StripTags))
			if sz+1 > table.ColumnWidths[c] {
				table.ColumnWidths[c] = sz
			}
			table.TrueCellSize[r][c] = sz
		}
	}

	if table.formatRowCount > 0 {
		var formatRowCols int
		for i := 0; i < table.formatRowCount; i++ {

			formatRowCols = len(table.Formatting[i])

			// Make sure there are enough formatting entries
			if formatRowCols < hdrColCt {

				for j := formatRowCols; j < hdrColCt; j++ {
					table.Formatting[j] = append(table.Formatting[j], `%s`)
				}

			}

		}

	}

	return table
}

func AnsiParse(input string) string {
	ansiLock.RLock()
	defer ansiLock.RUnlock()

	if forceAnsiFlags == AnsiTagsDefault {
		return input
	}

	if forceAnsiFlags == AnsiTagsParse {
		return ansitags.Parse(input)
	}

	if forceAnsiFlags == AnsiTagsStrip {
		return ansitags.Parse(input, ansitags.StripTags)
	}

	if forceAnsiFlags == AnsiTagsMono {
		return ansitags.Parse(input, ansitags.Monochrome)
	}

	return ansitags.Parse(input)
}

// Loads the ansi aliases from the config file
// Only if the file has been modified since the last load
func LoadAliases() {

	// Get the file info
	fInfo, err := os.Stat(util.FilePath(string(configs.GetFilePathsConfig().FolderDataFiles) + `/ansi-aliases.yaml`))
	// check if filemtime is not ansiAliasFileModTime
	if err != nil || fInfo.ModTime() == ansiAliasFileModTime {
		return
	}

	// Set to 256 color mode
	ansitags.SetColorMode(ansitags.Color256)

	ansiLock.Lock()
	defer ansiLock.Unlock()

	start := time.Now()

	ansiAliasFileModTime = fInfo.ModTime()
	if err = ansitags.LoadAliases(util.FilePath(string(configs.GetFilePathsConfig().FolderDataFiles) + `/ansi-aliases.yaml`)); err != nil {
		mudlog.Info("ansitags.LoadAliases()", "changed", true, "Time Taken", time.Since(start), "error", err.Error())
	}

	mudlog.Info("ansitags.LoadAliases()", "changed", true, "Time Taken", time.Since(start))
}
