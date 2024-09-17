package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Killstats(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	otherSuggestions := []string{}

	var headers []string

	tableTitle := `Kill Stats`

	rows := [][]string{}

	formatting := []string{
		`<ansi fg="mobname">%s</ansi>`,
		`<ansi fg="red">%s</ansi>`,
		`<ansi fg="230">%s</ansi>`,
	}
	totalKills := 0

	if rest == `race` || rest == `races` {

		tableTitle += ` by Race`

		headers = []string{`Race Name`, `Quantity`, `%`}

		totalKills = user.Character.KD.GetRaceKills()

		for raceName, killCt := range user.Character.KD.RaceKills {
			rows = append(rows, []string{
				raceName,
				fmt.Sprintf("%d", killCt),
				fmt.Sprintf("%2.f%%", float64(killCt)/float64(totalKills)*100),
			})
		}

		rows = append(rows, []string{
			``,
			``,
			``,
		})

		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats area</ansi>`)

	} else if rest == `zone` || rest == `zones` || rest == `area` || rest == `areas` {

		tableTitle += ` by Area`

		headers = []string{`Area Name`, `Quantity`, `%`}

		totalKills = user.Character.KD.GetRaceKills()

		for zoneName, killCt := range user.Character.KD.ZoneKills {
			rows = append(rows, []string{
				zoneName,
				fmt.Sprintf("%d", killCt),
				fmt.Sprintf("%2.f%%", float64(killCt)/float64(totalKills)*100),
			})
		}

		rows = append(rows, []string{
			``,
			``,
			``,
		})

		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats race</ansi>`)

	} else {

		tableTitle += ` by Mob`

		headers = []string{`Mob Name`, `Quantity`, `%`}

		totalKills = user.Character.KD.GetMobKills()

		for mobId, killCt := range user.Character.KD.Kills {
			if mobSpec := mobs.GetMobSpec(mobs.MobId(mobId)); mobSpec != nil {

				rows = append(rows, []string{
					mobSpec.Character.Name,
					fmt.Sprintf("%d", killCt),
					fmt.Sprintf("%2.f%%", float64(killCt)/float64(totalKills)*100),
				})
			}
		}

		rows = append(rows, []string{
			``,
			``,
			``,
		})

		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats area</ansi>`)
		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats race</ansi>`)
	}

	rows = append(rows, []string{
		`Total Kills`,
		fmt.Sprintf("%d", totalKills),
		``,
	})

	if user.Character.KD.GetDeaths() == 0 {
		rows = append(rows, []string{
			`Total Deaths`,
			fmt.Sprintf("%d", user.Character.KD.GetDeaths()),
			`N/A`,
		})
	} else {
		rows = append(rows, []string{
			`Total Deaths`,
			fmt.Sprintf("%d", user.Character.KD.GetDeaths()),
			fmt.Sprintf("%.2f:1", user.Character.KD.GetKDRatio()),
		})
	}

	searchResultsTable := templates.GetTable(tableTitle, headers, rows, formatting)
	tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
	tplTxt += fmt.Sprintf("Also try: %s\n", strings.Join(otherSuggestions, `, `))
	user.SendText(tplTxt)

	//user.SendText(fmt.Sprintf(`Also try: %s`, strings.Join(otherSuggestions, `, `)))

	return true, nil
}
