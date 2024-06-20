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

	headers := []string{`SpellId`, `Name`, `Description`, `Target`, `Cost`, `Rounds`}
	allFormatting := [][]string{}
	rows := [][]string{}

	for _, spellId := range user.Character.GetSpells() {

		allFormatting = append(allFormatting, []string{
			`<ansi fg="yellow-bold">%s</ansi>`,
			`<ansi fg="yellow-bold">%s</ansi>`,
			`<ansi fg="yellow-bold">%s</ansi>`,
			`<ansi fg="white-bold">%s</ansi>`,
			`<ansi fg="magenta-bold">%s</ansi>`,
			`<ansi fg="mana-bold">%s</ansi>`,
		})

		if sp := spells.GetSpell(spellId); sp != nil {

			target := string(sp.Type)
			switch sp.Type {
			case spells.Neutral:
				target = `?`
			case spells.HelpSingle:
				target = `Single`
			case spells.HarmSingle:
				target = `Single`
			case spells.HelpMulti:
				target = `Multi`
			case spells.HarmMulti:
				target = `Multi`
			}
			rows = append(rows, []string{sp.SpellId, sp.Name, sp.Description, target, fmt.Sprintf(`%d`, sp.Cost), fmt.Sprintf(`%d`, sp.WaitRounds)})

		}

	}

	onlineResultsTable := templates.GetTable(`Spells`, headers, rows, allFormatting...)
	tplTxt, _ := templates.Process("tables/generic", onlineResultsTable)
	response.SendUserMessage(userId, tplTxt, false)

	response.Handled = true
	return response, nil
}
