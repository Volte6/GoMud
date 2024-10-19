package usercommands

import (
	"strings"

	"github.com/volte6/gomud/items"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
)

func Reload(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.reload", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	switch strings.ToLower(rest) {
	case `items`:
		items.LoadDataFiles()
		user.SendText(`Items reloaded.`)
	default:
		user.SendText(`Unknown reload command.`)
	}
	return true, nil
}
