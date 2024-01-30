package usercommands

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

var (
	memoryReportCache = map[string]util.MemoryResult{}
)

func Server(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.server", nil)
		response.SendUserMessage(userId, infoOutput, false)
		response.Handled = true
		return response, nil
	}

	args := util.SplitButRespectQuotes(rest)
	if args[0] == "set" {

		args = args[1:]

		if len(args) < 2 {

			headers := []string{"Name", "Value"}
			rows := [][]string{}
			formatting := []string{`<ansi fg="yellow" bold="true">%s</ansi>`, `<ansi fg="red" bold="true">%s</ansi>`}

			response.SendUserMessage(userId, ``, true)

			cfgData := configs.GetConfig().AllConfigData()
			cfgKeys := make([]string, 0, len(cfgData))
			for k := range cfgData {
				cfgKeys = append(cfgKeys, k)
			}

			// sort the keys
			slices.Sort(cfgKeys)

			for _, k := range cfgKeys {
				rows = append(rows, []string{k, fmt.Sprintf(`%v`, cfgData[k])})
			}

			settingsTable := templates.GetTable("Server Settings", headers, rows, formatting)
			tplTxt, _ := templates.Process("tables/generic", settingsTable)
			response.SendUserMessage(userId, tplTxt, true)

			response.Handled = true
			return response, nil
		}

		configName := strings.ToLower(args[0])
		configValue := strings.Join(args[1:], ` `)

		if err := configs.SetVal(configName, configValue); err != nil {
			response.SendUserMessage(userId, fmt.Sprintf(`config change error: %s=%s (%s)`, configName, configValue, err), true)
			response.Handled = true
			return response, nil
		}

		response.SendUserMessage(userId, fmt.Sprintf(`config changed: %s=%s`, configName, configValue), true)

		response.Handled = true
		return response, nil
	}

	if rest == "reload-ansi" {
		templates.LoadAliases()
		response.SendUserMessage(userId, `ansi aliases reloaded`, true)
		response.Handled = true
		return response, nil
	}

	if rest == "ansi-passthrough" {
		templates.SetAnsiFlag(templates.AnsiTagsIgnore)
	}

	if rest == "ansi-strip" {
		templates.SetAnsiFlag(templates.AnsiTagsStrip)
	}

	if rest == "ansi-mono" {
		templates.SetAnsiFlag(templates.AnsiTagsMono)
	}

	if rest == "ansi-preparse" {
		templates.SetAnsiFlag(templates.AnsiTagsPreParse)
	}

	if rest == "ansi-normal" {
		templates.SetAnsiFlag(templates.AnsiTagsDefault)
	}

	if rest == "stats" || rest == "info" {

		//
		// General Go stats
		//
		response.SendUserMessage(userId, ``, true)
		response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="yellow" bold="true">IP/Port:</ansi>    <ansi fg="red">%s</ansi>`, util.GetServerAddress()), true)
		response.SendUserMessage(userId, ``, true)

		//
		// Special timing related stats
		//
		headers := []string{"Routine", "Avg", "Low", "High", "Ct", "/sec"}
		rows := [][]string{}
		formatting := []string{`<ansi fg="yellow" bold="true">%s</ansi>`, `<ansi fg="cyan" bold="true">%s</ansi>`, `<ansi fg="cyan" bold="true">%s</ansi>`, `<ansi fg="cyan" bold="true">%s</ansi>`, `<ansi fg="black" bold="true">%s</ansi>`, `<ansi fg="black" bold="true">%s</ansi>`}

		allTimers := map[string]util.Accumulator{}
		allNames := []string{}

		times := util.GetTimeTrackers()
		for _, timeAcc := range times {

			allNames = append(allNames, timeAcc.Name)
			allTimers[timeAcc.Name] = timeAcc
		}

		sort.Strings(allNames)
		for _, name := range allNames {
			acc := allTimers[name]
			lowest, highest, average, ct := acc.Stats()
			rows = append(rows, []string{acc.Name,
				fmt.Sprintf(`%4.3fms`, average*1000),
				fmt.Sprintf(`%4.3fms`, lowest*1000),
				fmt.Sprintf(`%4.3fms`, highest*1000),
				fmt.Sprintf(`%d`, int(ct)),
				fmt.Sprintf(`%4.3f`, ct/time.Since(acc.Start).Seconds()),
			})
		}

		tblData := templates.GetTable(`Timer Stats`, headers, rows, formatting)
		tplTxt, _ := templates.Process("tables/generic", tblData)
		response.SendUserMessage(userId, tplTxt, true)

		//
		// Alternative rendering
		//
		memRepHeaders := []string{"Section  ", "Items    ", "KB       ", "MB       ", "GB       ", "Change   "}
		memRepFormatting := []string{`<ansi fg="yellow" bold="true">%s</ansi>`,
			`<ansi fg="black" bold="true">%s</ansi>`,
			`<ansi fg="cyan" bold="true">%s</ansi>`,
			`<ansi fg="red">%s</ansi>`,
			`<ansi fg="red" bold="true">%s</ansi>`,
			`<ansi fg="black" bold="true">%s</ansi>`}

		memRepRows := [][]string{}
		memRepTotalTotal := uint64(0)

		sectionNames, memReports := util.GetMemoryReport()

		for idx, memReport := range memReports {

			sectionName := sectionNames[idx]

			tmpRowStorage := map[string]util.MemoryResult{}
			var memRepRowNames []string = []string{}
			var memRepTotal uint64 = 0

			for name, memResult := range memReport {
				usage := memResult.Memory
				memRepRowNames = append(memRepRowNames, name)
				tmpRowStorage[name] = memResult
				memRepTotal += usage
			}

			memRepRows = append(memRepRows, []string{`[ ` + sectionName + ` ]`, ``, ``, ``, ``, ``})
			sort.Strings(memRepRowNames)
			for _, name := range memRepRowNames {

				var rowData []string

				var prevString string = ``
				var prevCtString string = ``
				if cachedMemResult, ok := memoryReportCache[name]; ok {
					val := cachedMemResult.Memory
					if val > tmpRowStorage[name].Memory { // It has gone down
						prevString = fmt.Sprintf(`↓%s`, util.FormatBytes(val-tmpRowStorage[name].Memory))
					} else if val < tmpRowStorage[name].Memory {
						prevString = fmt.Sprintf(`↑%s`, util.FormatBytes(tmpRowStorage[name].Memory-val))
					}

					ct := cachedMemResult.Count
					if ct > tmpRowStorage[name].Count { // It has gone down
						prevCtString = fmt.Sprintf(`(↓%d)`, ct-tmpRowStorage[name].Count)
					} else if ct < tmpRowStorage[name].Count {
						prevCtString = fmt.Sprintf(`(↑%d)`, tmpRowStorage[name].Count-ct)
					}
				}
				memoryReportCache[name] = tmpRowStorage[name] // Cache the new val

				// foramt the new val
				bFormatted := util.FormatBytes(tmpRowStorage[name].Memory)

				count := ``
				if tmpRowStorage[name].Count > 0 {
					count = fmt.Sprintf(`%d %s`, tmpRowStorage[name].Count, prevCtString)
				}
				if strings.Contains(bFormatted, `KB`) {
					rowData = []string{name, count, bFormatted, ``, ``, prevString}
				} else if strings.Contains(bFormatted, `MB`) {
					rowData = []string{name, count, ``, bFormatted, ``, prevString}
				} else if strings.Contains(bFormatted, `GB`) {
					rowData = []string{name, count, ``, ``, bFormatted, prevString}
				} else {
					rowData = []string{name, count, ``, ``, ``, prevString}
				}

				memRepRows = append(memRepRows, rowData)
			}
			memRepRows = append(memRepRows, []string{``, ``, ``, ``, ``, ``})

			if sectionName != `Go` {
				memRepTotalTotal += memRepTotal
			}
			memRepTotal = 0
		}

		var rowData []string

		var name string = `Total (Non Go)`
		var prevString string = ``
		if cachedMemResult, ok := memoryReportCache[name]; ok {
			val := cachedMemResult.Memory
			if val > memRepTotalTotal { // It has gone down
				prevString = fmt.Sprintf(`↓%s`, util.FormatBytes(val-memRepTotalTotal))
			} else if val < memRepTotalTotal {
				prevString = fmt.Sprintf(`↑%s`, util.FormatBytes(memRepTotalTotal-val))
			}
		}

		memoryReportCache[name] = util.MemoryResult{memRepTotalTotal, 0} // Cache the new val

		bFormatted := util.FormatBytes(memRepTotalTotal)
		if strings.Contains(bFormatted, `KB`) {
			rowData = []string{`Total (Non Go)`, ``, bFormatted, ``, ``, prevString}
		} else if strings.Contains(bFormatted, `MB`) {
			rowData = []string{`Total (Non Go)`, ``, ``, bFormatted, ``, prevString}
		} else if strings.Contains(bFormatted, `GB`) {
			rowData = []string{`Total (Non Go)`, ``, ``, ``, bFormatted, prevString}
		} else {
			rowData = []string{`Total (Non Go)`, ``, ``, ``, ``, prevString}
		}

		memRepRows = append(memRepRows, rowData)
		memRepTblData := templates.GetTable(`Specific Memory`, memRepHeaders, memRepRows, memRepFormatting)
		memRepTxt, _ := templates.Process("tables/generic", memRepTblData)
		response.SendUserMessage(userId, memRepTxt, true)
	}

	response.Handled = true
	return response, nil
}
