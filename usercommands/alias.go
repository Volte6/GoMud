package usercommands

import (
	"fmt"
	"sort"

	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Alias(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// biuld array and look up table for sorting purposes
	allOutCmds := []string{}
	reverseLookup := map[string][]string{}

	for inCmd, outCmd := range keywords.GetAllCommandAliases() {
		if _, ok := reverseLookup[outCmd]; !ok {
			reverseLookup[outCmd] = []string{}
			allOutCmds = append(allOutCmds, outCmd)
		}
		reverseLookup[outCmd] = append(reverseLookup[outCmd], inCmd)
	}

	sort.Strings(allOutCmds)

	headers := []string{"Alias", "Command"}
	rows := [][]string{}

	user.SendText(`<ansi fg="yellow">Built in Aliases:</ansi>`)

	for _, outCmd := range allOutCmds {
		inCmds := reverseLookup[outCmd]
		for _, inCmd := range inCmds {
			rows = append(rows, []string{inCmd, outCmd})
		}
	}

	tableFormatting := []string{`<ansi fg="yellow">%s</ansi>`, `<ansi fg="command">%s</ansi>`}

	aliasTableData := templates.GetTable(`Default Aliases`, headers, rows, tableFormatting)
	aliasTxt, _ := templates.Process("tables/generic", aliasTableData)
	user.SendText(aliasTxt)

	response.Handled = true
	return response, nil
}
