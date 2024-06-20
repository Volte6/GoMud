package usercommands

import (
	"fmt"

	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Spells(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
	}

	headers := []string{`SpellId`, `Name`, `Description`, `Target`, `Cost`, `Rounds`, `Casts`}
	allFormatting := [][]string{}
	rows := [][]string{}

	for spellId, casts := range user.Character.GetSpells() {

		if casts < 0 {
			continue
		}

		casts -= 1

		if sp := spells.GetSpell(spellId); sp != nil {

			targetColor := `green-bold`
			target := string(sp.Type)
			switch sp.Type {
			case spells.Neutral:
				target = `?`
				targetColor = `white-bold`
			case spells.HelpSingle:
				target = `Single`
			case spells.HarmSingle:
				target = `Single`
				targetColor = `red-bold`
			case spells.HelpMulti:
				target = `Multi`
			case spells.HarmMulti:
				target = `Multi`
				targetColor = `red-bold`
			}

			allFormatting = append(allFormatting, []string{
				`<ansi fg="yellow-bold">%s</ansi>`,
				`<ansi fg="yellow-bold">%s</ansi>`,
				`<ansi fg="yellow-bold">%s</ansi>`,
				`<ansi fg="` + targetColor + `">%s</ansi>`,
				`<ansi fg="magenta-bold">%s</ansi>`,
				`<ansi fg="mana-bold">%s</ansi>`,
				`<ansi fg="red-bold">%s</ansi>`,
			})

			rows = append(rows, []string{sp.SpellId, sp.Name, sp.Description, target, fmt.Sprintf(`%d`, sp.Cost), fmt.Sprintf(`%d`, sp.WaitRounds), fmt.Sprintf(`%d`, casts)})

		}

	}

	onlineResultsTable := templates.GetTable(`Spells`, headers, rows, allFormatting...)
	tplTxt, _ := templates.Process("tables/generic", onlineResultsTable)
	response.SendUserMessage(userId, tplTxt, false)

	response.Handled = true
	return response, nil
}
