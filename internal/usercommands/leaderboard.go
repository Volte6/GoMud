package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/leaderboard"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Leaderboards(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if configs.GetConfig().LeaderboardSize == 0 {
		user.SendText(`Leaderboards are disabled.`)
		return true, nil
	}

	allLeaderboards := leaderboard.Get()

	for lbName, entries := range allLeaderboards {

		valueName := strings.Title(lbName)
		title := fmt.Sprintf(`%s Leaderboard`, valueName)

		headers := []string{`Rank`, `Character`, `Profession`, `Level`, valueName}

		rows := [][]string{}

		formatting := []string{
			`<ansi fg="red">%s</ansi>`,
			`<ansi fg="username">%s</ansi>`,
			`<ansi fg="white-bold">%s</ansi>`,
			`<ansi fg="157">%s</ansi>`,
		}

		if lbName == "experience" {
			formatting = append(formatting, `<ansi fg="experience">%s</ansi>`)
		} else if lbName == "gold" {
			formatting = append(formatting, `<ansi fg="gold">%s</ansi>`)
		} else if lbName == "kills" {
			formatting = append(formatting, `<ansi fg="red">%s</ansi>`)
		}

		for i, entry := range entries {

			if entry.CharacterName == `` {
				continue
			}

			newRow := []string{`#` + strconv.Itoa(i+1), entry.CharacterName, entry.CharacterClass, strconv.Itoa(entry.Level)}

			if lbName == "experience" {
				newRow = append(newRow, strconv.Itoa(entry.Experience))
			} else if lbName == "gold" {
				newRow = append(newRow, strconv.Itoa(entry.Gold))
			} else if lbName == "kills" {
				newRow = append(newRow, strconv.Itoa(entry.Kills))
			}

			rows = append(rows, newRow)
		}

		searchResultsTable := templates.GetTable(title, headers, rows, formatting)
		tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
		user.SendText(tplTxt)

	}
	return true, nil
}
