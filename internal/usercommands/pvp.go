package usercommands

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Pvp(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	setting := configs.GetConfig().PVP

	user.SendText("")
	if setting == configs.PVPDisabled {
		user.SendText(`PVP is <ansi fg="alert-5">disabled</ansi> on this server. You cannot fight other players.`)
	} else if setting == configs.PVPEnabled {
		user.SendText(`PVP is <ansi fg="green-bold">enabled</ansi> on this server. You can fight other players anywhere.`)
	} else if setting == configs.PVPLimited {
		user.SendText(`PVP is <ansi fg="yellow">limited</ansi> on this server. You can fight other players in places labeled with: <ansi fg="11" bg="52"> ☠ PK Area ☠ </ansi>.`)
	}
	user.SendText("")

	return true, nil
}
