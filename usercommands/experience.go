package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Experience(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) > 0 && args[0] == `chart` {

		args = args[1:]

		startLevel := 1
		endLevel := 25

		chartRace := user.Character.RaceId

		// xp chart elf 50
		if len(args) > 1 {

			if lvl, err := strconv.Atoi(args[len(args)-1]); err == nil {
				endLevel = lvl
				args = args[:len(args)-1]
			} else if lvl, err := strconv.Atoi(args[0]); err == nil {
				endLevel = lvl
				args = args[1:]
			}

			raceName := strings.Join(args, ` `)
			if raceInfo, found := races.FindRace(raceName); found {
				chartRace = raceInfo.RaceId
			}
		} else if len(args) == 1 {

			if lvl, err := strconv.Atoi(args[0]); err == nil {
				endLevel = lvl
			} else {
				if raceInfo, found := races.FindRace(args[0]); found {
					chartRace = raceInfo.RaceId
				}
			}

		}

		if endLevel < 2 {
			endLevel = 2
		}
		if endLevel > 200 {
			endLevel = 200
		}
		if endLevel-startLevel > 25 {
			startLevel = endLevel - 25
		}
		if startLevel < 1 {
			startLevel = 1
		}

		mockChar := characters.New()
		mockChar.RaceId = chartRace
		mockChar.Validate()

		headers := []string{`Level`, `Experience`, `Str`, `Spd`, `Smt`, `Vit`, `Mys`, `Per`, `ALL`}
		rows := [][]string{}

		formatting := []string{
			`<ansi fg="white-bold">%s</ansi>`,
			`<ansi fg="red">%s</ansi>`,
			`<ansi fg="yellow">%s</ansi>`,
			`<ansi fg="yellow">%s</ansi>`,
			`<ansi fg="yellow">%s</ansi>`,
			`<ansi fg="yellow">%s</ansi>`,
			`<ansi fg="yellow">%s</ansi>`,
			`<ansi fg="yellow">%s</ansi>`,
			`<ansi fg="white">%s</ansi>`,
		}

		zeroStr := ``

		stats := []string{`str`, `spd`, `smt`, `vit`, `mys`, `per`}

		oldG := map[string]int{
			`str`: mockChar.Stats.Strength.GainsForLevel(startLevel - 1),
			`spd`: mockChar.Stats.Speed.GainsForLevel(startLevel - 1),
			`smt`: mockChar.Stats.Smarts.GainsForLevel(startLevel - 1),
			`vit`: mockChar.Stats.Vitality.GainsForLevel(startLevel - 1),
			`mys`: mockChar.Stats.Mysticism.GainsForLevel(startLevel - 1),
			`per`: mockChar.Stats.Perception.GainsForLevel(startLevel - 1),
		}

		newG := map[string]int{}
		totalG := map[string]int{}
		for _, stat := range stats {
			totalG[stat] = oldG[stat]
		}

		for i := startLevel; i <= endLevel; i++ {

			newG = map[string]int{
				`str`: mockChar.Stats.Strength.GainsForLevel(i),
				`spd`: mockChar.Stats.Speed.GainsForLevel(i),
				`smt`: mockChar.Stats.Smarts.GainsForLevel(i),
				`vit`: mockChar.Stats.Vitality.GainsForLevel(i),
				`mys`: mockChar.Stats.Mysticism.GainsForLevel(i),
				`per`: mockChar.Stats.Perception.GainsForLevel(i),
			}

			tnlXP := mockChar.XPTL(i) - mockChar.XPTL(i-1)

			if i == 1 {
				tnlXP = 0
			}

			if i > 1 {

				row := []string{fmt.Sprintf(`%d`, i), fmt.Sprintf(`%d`, tnlXP)}
				gainStr := zeroStr
				all := 0
				for _, stat := range stats {
					gain := newG[stat] - oldG[stat]
					all += gain
					totalG[stat] += gain

					if gain > 0 {
						gainStr = fmt.Sprintf(`%d`, gain)
					} else {
						gainStr = zeroStr
					}

					row = append(row, gainStr)
				}

				if all > 0 {
					row = append(row, fmt.Sprintf(`%d`, all))
				} else {
					row = append(row, zeroStr)
				}

				rows = append(rows, row)
			}

			for _, stat := range stats {
				oldG[stat] = newG[stat]
			}
		}

		row := []string{`Total`, fmt.Sprintf(`%d`, mockChar.XPTL(endLevel)-1500)}
		all := 0
		for _, stat := range stats {
			gain := totalG[stat]
			all += gain
			gainStr := ``
			if gain > 0 {
				gainStr = fmt.Sprintf(`%d`, gain)
			}

			row = append(row, gainStr)
		}

		if all > 0 {
			row = append(row, fmt.Sprintf(`%d`, all))
		} else {
			row = append(row, zeroStr)
		}

		rows = append(rows, row)

		raceInfo := races.GetRace(mockChar.RaceId)
		searchResultsTable := templates.GetTable(fmt.Sprintf(`Experience Chart for %s`, raceInfo.Name), headers, rows, formatting)
		tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
		response.SendUserMessage(userId, tplTxt, false)

		response.Handled = true
		return response, nil
	}

	realXPNow, realXPTNL := user.Character.XPTNLActual()
	xpInfo := map[string]int{
		"Level": user.Character.Level,
		"Exp":   realXPNow,
		"Tnl":   realXPTNL,
		"Tp":    user.Character.TrainingPoints,
		"Sp":    user.Character.StatPoints,
	}

	tplTxt, _ := templates.Process("character/experience", xpInfo)
	response.SendUserMessage(userId, tplTxt, false)

	response.Handled = true
	return response, nil
}
