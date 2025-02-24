package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/badinputtracker"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func BadCommands(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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
