package usercommands

import (
	"fmt"

	"github.com/volte6/mud/badinputtracker"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func BadCommands(rest string, user *users.UserRecord) (bool, error) {

	if rest == "clear" {
		badinputtracker.Clear()
	}

	// Now look for mobs?
	headers := []string{"Command", "Count"}
	rows := [][]string{}

	for cmd, ct := range badinputtracker.GetBadCommands() {
		rows = append(rows, []string{
			cmd,
			fmt.Sprintf(`%d`, ct),
		})
	}

	badCommandTableData := templates.GetTable(`Bad Commands`, headers, rows)
	tplTxt, _ := templates.Process("tables/generic", badCommandTableData)

	user.SendText(tplTxt)

	return true, nil
}
