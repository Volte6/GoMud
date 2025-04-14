package usercommands

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

/*
* Role Permissions:
* reload 				(All)
 */
func Reload(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.reload", nil, user.UserId)
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
