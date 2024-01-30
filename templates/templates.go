package templates

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"log/slog"

	"github.com/Volte6/ansitags"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/util"
)

type AnsiFlag uint8

const (
	AnsiTagsDefault  AnsiFlag = iota // Parse ansi tags, and use default color information
	AnsiTagsIgnore                   // Do nothing, even if tags exist
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
	forceAnsiFlags       = AnsiTagsDefault
	ansiLock             sync.RWMutex
	ansiAliasFileModTime time.Time
)

func Exists(name string) bool {

	var fullPath string = util.FilePath(configs.GetConfig().FolderTemplates, `/`, name+`.template`)
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

	var ignoreAnsiTags bool = false
	var preParseAnsiTags bool = false

	var ansitagsParseBehavior []ansitags.ParseBehavior = make([]ansitags.ParseBehavior, 0, 5)

	if forceAnsiFlags != AnsiTagsDefault {
		ansiFlags = append(ansiFlags, forceAnsiFlags)
	}

	for _, flag := range ansiFlags {
		switch flag {
		case AnsiTagsStrip:
			ansitagsParseBehavior = append(ansitagsParseBehavior, ansitags.StripTags)
		case AnsiTagsMono:
			ansitagsParseBehavior = append(ansitagsParseBehavior, ansitags.Monochrome)
		case AnsiTagsIgnore:
			ignoreAnsiTags = true
		case AnsiTagsPreParse:
			preParseAnsiTags = true
		}
	}

	// All templates must end with .template
	var fullPath string = util.FilePath(configs.GetConfig().FolderTemplates, `/`, name+`.template`)

	fInfo, err := os.Stat(fullPath)
	if err != nil {
		//slog.Error("could not stat template file", "error", err)
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
			slog.Error("could not read template file", "error", err)
			return "[TEMPLATE READ ERROR]", err
		}

		if !ignoreAnsiTags && preParseAnsiTags {
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
		slog.Error("could not parse template file", "error", err)
		return "[TEMPLATE ERROR]", err
	}

	// return the final data as a string, parse ansi tags if needed (No need to parse if it was preparsed)
	if !ignoreAnsiTags && !cache.ansiPreparsed {
		return ansitags.Parse(buf.String(), ansitagsParseBehavior...), nil
	}

	return buf.String(), nil
}

const cellPadding int = 1

type TemplateTable struct {
	Title          string
	Header         []string
	Rows           [][]string
	ColumnCount    int
	ColumnWidths   []int
	Formatting     [][]string
	formatRowCount int
}

func (t TemplateTable) GetCell(row int, column int) string {

	cellStr := t.Rows[row][column]
	repeatCt := t.ColumnWidths[column] - len([]rune(cellStr))
	if repeatCt > 0 {
		cellStr += strings.Repeat(` `, repeatCt)
	}

	if t.formatRowCount > 0 {
		return fmt.Sprintf(t.Formatting[row%t.formatRowCount][column], cellStr)
	}
	return cellStr
}

func GetTable(title string, headers []string, rows [][]string, formatting ...[]string) TemplateTable {

	var table TemplateTable = TemplateTable{
		Title:        title,
		Header:       headers,
		Rows:         rows,
		ColumnCount:  len(headers),
		ColumnWidths: make([]int, len(headers)),
		Formatting:   formatting,
	}

	hdrColCt := len(headers)
	rowCt := len(rows)
	table.formatRowCount = len(formatting)

	// Get the longest element
	for i := 0; i < hdrColCt; i++ {
		if len(headers[i])+1 > table.ColumnWidths[i] {
			table.ColumnWidths[i] = len([]rune(headers[i]))
		}
	}

	// Get the longest element
	for r := 0; r < rowCt; r++ {

		rowColCt := len(rows[r])
		if hdrColCt < rowColCt {
			for i := hdrColCt; i < rowColCt; i++ {
				table.Header = append(table.Header, ``)
			}
			hdrColCt = len(table.Header)
		}

		for c := 0; c < hdrColCt; c++ {
			if len(rows[r][c])+1 > table.ColumnWidths[c] {
				table.ColumnWidths[c] = len([]rune(rows[r][c]))
			}
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
		return ansitags.Parse(input)
	}

	if forceAnsiFlags == AnsiTagsIgnore {
		return input
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
	fInfo, err := os.Stat(util.FilePath(configs.GetConfig().FileAnsiAliases))
	// check if filemtime is not ansiAliasFileModTime
	if err != nil || fInfo.ModTime() == ansiAliasFileModTime {
		return
	}

	ansiLock.Lock()
	defer ansiLock.Unlock()

	start := time.Now()

	ansiAliasFileModTime = fInfo.ModTime()
	if err = ansitags.LoadAliases(util.FilePath(configs.GetConfig().FileAnsiAliases)); err != nil {
		slog.Info("ansitags.LoadAliases()", "changed", true, "Time Taken", time.Since(start), "error", err.Error())
	}

	slog.Info("ansitags.LoadAliases()", "changed", true, "Time Taken", time.Since(start))
}
