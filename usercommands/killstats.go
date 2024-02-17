package usercommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Killstats(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	headers := []string{`Mob Name`, `Quantity`, `%`}
	rows := [][]string{}

	formatting := []string{
		`<ansi fg="mobname">%s</ansi>`,
		`<ansi fg="red">%s</ansi>`,
		`<ansi fg="230">%s</ansi>`,
	}

	totalKills := user.Character.KD.GetKills()

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

	searchResultsTable := templates.GetTable(`Kill Stats`, headers, rows, formatting)
	tplTxt, _ := templates.Process("tables/generic", searchResultsTable)
	response.SendUserMessage(userId, tplTxt, false)

	response.Handled = true
	return response, nil
}
