package usercommands

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/volte6/gomud/configs"
	"github.com/volte6/gomud/gametime"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
	"github.com/volte6/gomud/util"
)

var (
	memoryReportCache = map[string]util.MemoryResult{}
)

func Server(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.server", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	args := util.SplitButRespectQuotes(rest)
	if args[0] == "set" {

		args = args[1:]

		if len(args) < 1 {

			headers := []string{"Name", "Value"}
			rows := [][]string{}
			formatting := []string{`<ansi fg="yellow-bold">%s</ansi>`, `<ansi fg="red-bold">%s</ansi>`}

			user.SendText(``)

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
			user.SendText(tplTxt)

			return true, nil
		}

		if args[0] == "day" {
			gametime.SetToDay(-1)
			gd := gametime.GetDate()
			user.SendText(`Time set to ` + gd.String())
			return true, nil
		} else if args[0] == "night" {
			gametime.SetToNight(-1)
			gd := gametime.GetDate()
			user.SendText(`Time set to ` + gd.String())
			return true, nil
		} else if args[0] == "time" && len(args) > 1 {

			timeStr := strings.Join(args[1:], ` `)

			if len(timeStr) >= 2 && strings.ToLower(timeStr[len(timeStr)-2:]) == `pm` {
				timeStr = timeStr[:len(timeStr)-2]
			}

			timeParts := strings.Split(timeStr, `:`)

			hourStr := timeParts[0]
			minuteStr := `0`
			if len(timeParts) > 1 {
				minuteStr = timeParts[1]
			}

			hour, _ := strconv.Atoi(hourStr)
			minutes, _ := strconv.Atoi(minuteStr)

			gametime.SetTime(hour, minutes)
			gd := gametime.GetDate()
			user.SendText(`Time set to ` + gd.String())
			return true, nil
		}

		configName := strings.ToLower(args[0])
		configValue := strings.Join(args[1:], ` `)

		if err := configs.SetVal(configName, configValue); err != nil {
			user.SendText(fmt.Sprintf(`config change error: %s=%s (%s)`, configName, configValue, err))
			return true, nil
		}

		user.SendText(fmt.Sprintf(`config changed: %s=%s`, configName, configValue))

		return true, nil
	}

	if rest == "reload-ansi" {
		templates.LoadAliases()
		user.SendText(`ansi aliases reloaded`)
		return true, nil
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
		user.SendText(``)
		user.SendText(fmt.Sprintf(`<ansi fg="yellow-bold">IP/Port:</ansi>    <ansi fg="red">%s</ansi>`, util.GetServerAddress()))
		user.SendText(``)

		//
		// Special timing related stats
		//
		headers := []string{"Routine", "Avg", "Low", "High", "Ct", "/sec"}
		rows := [][]string{}
		formatting := []string{`<ansi fg="yellow-bold">%s</ansi>`, `<ansi fg="cyan-bold">%s</ansi>`, `<ansi fg="cyan-bold">%s</ansi>`, `<ansi fg="cyan-bold">%s</ansi>`, `<ansi fg="black-bold">%s</ansi>`, `<ansi fg="black-bold">%s</ansi>`}

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
		user.SendText(tplTxt)

		//
		// Alternative rendering
		//
		memRepHeaders := []string{"Section  ", "Items    ", "KB       ", "MB       ", "GB       ", "Change   "}
		memRepFormatting := []string{`<ansi fg="yellow-bold">%s</ansi>`,
			`<ansi fg="black-bold">%s</ansi>`,
			`<ansi fg="cyan-bold">%s</ansi>`,
			`<ansi fg="red">%s</ansi>`,
			`<ansi fg="red-bold">%s</ansi>`,
			`<ansi fg="black-bold">%s</ansi>`}

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
		user.SendText(memRepTxt)
	}

	return true, nil
}
