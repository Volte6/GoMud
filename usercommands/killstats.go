package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
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

	mobKills := map[string]int{}
	raceKills := map[string]int{}
	areaKills := map[string]int{}

	for mid, kCt := range user.Character.KD.Kills {

		if mobSpec := mobs.GetMobSpec(mobs.MobId(mid)); mobSpec != nil {

			totalKills += kCt

			// Populate mob kills
			mobKills[mobSpec.Character.Name] = mobKills[mobSpec.Character.Name] + kCt

			// Populate race kills
			if raceInfo := races.GetRace(mobSpec.Character.RaceId); raceInfo != nil {
				raceKills[raceInfo.Name] = raceKills[raceInfo.Name] + kCt
			}

			// Populate area kills
			areaKills[mobSpec.Zone] = areaKills[mobSpec.Zone] + kCt
		}
	}

	renderStats := mobKills

	if rest == `race` || rest == `races` {

		renderStats = raceKills
		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats area</ansi>`)

	} else if rest == `zone` || rest == `zones` || rest == `area` || rest == `areas` {

		renderStats = areaKills
		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats race</ansi>`)

	} else {

		rest = `mob` // default to mob

		renderStats = mobKills
		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats area</ansi>`)
		otherSuggestions = append(otherSuggestions, `<ansi fg="command">killstats race</ansi>`)
	}

	headers = []string{strings.Title(rest), `Quantity`, `%`}

	for name, killCt := range renderStats {

		rows = append(rows, []string{
			name,
			fmt.Sprintf("%d", killCt),
			fmt.Sprintf("%2.f%%", float64(killCt)/float64(totalKills)*100),
		})
	}

	rows = append(rows, []string{
		``,
		``,
		``,
	})

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

	searchResultsTable := templates.GetTable(tableTitle+` by `+strings.Title(rest), headers, rows, formatting)
	tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
	tplTxt += fmt.Sprintf("Also try: %s\n", strings.Join(otherSuggestions, `, `))
	user.SendText(tplTxt)

	//user.SendText(fmt.Sprintf(`Also try: %s`, strings.Join(otherSuggestions, `, `)))

	return true, nil
}
