package usercommands

import (
	"fmt"
	"strconv"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/leaderboard"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Leaderboards(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if configs.GetStatisticsConfig().Leaderboards.Size == 0 {
		user.SendText(`Leaderboards are disabled.`)
		return true, nil
	}

	leaderboard.Update()

	for _, lb := range leaderboard.Get() {

		title := fmt.Sprintf(`%s Leaderboard`, lb.Name)

		headers := []string{`Rank`, `Character`, `Profession`, `Level`, lb.Name}

		rows := [][]string{}

		valueFormatting := `%s`
		if lb.ValueColor != `` {
			valueFormatting = `<ansi fg="` + lb.ValueColor + `">%s</ansi>`
		}

		formatting := []string{
			`<ansi fg="red">%s</ansi>`,
			`<ansi fg="username">%s</ansi>`,
			`<ansi fg="white-bold">%s</ansi>`,
			`<ansi fg="157">%s</ansi>`,
			valueFormatting,
		}

		for i, entry := range lb.Top {

			if entry.UserId == 0 {
				continue
			}

			newRow := []string{`#` + strconv.Itoa(i+1), entry.CharacterName, entry.CharacterClass, strconv.Itoa(entry.Level), util.FormatNumber(entry.ScoreValue)}

			rows = append(rows, newRow)
		}

		searchResultsTable := templates.GetTable(title, headers, rows, formatting)
		tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
		user.SendText("\n")
		user.SendText(tplTxt)

	}
	return true, nil
}
