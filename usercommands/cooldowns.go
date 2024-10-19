package usercommands

import (
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
)

func Cooldowns(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	cdTxt, _ := templates.Process("character/cooldowns", user.Character.GetAllCooldowns())
	user.SendText(cdTxt)

	return true, nil
}
