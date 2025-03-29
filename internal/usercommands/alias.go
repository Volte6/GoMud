package usercommands

import (
	"sort"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/keywords"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Alias(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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
	aliasTxt, _ := templates.Process("tables/generic", aliasTableData, user.UserId)
	user.SendText(aliasTxt)

	return true, nil
}
