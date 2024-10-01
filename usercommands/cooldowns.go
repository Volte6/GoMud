package usercommands

import (
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Cooldowns(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	cdTxt, _ := templates.Process("character/cooldowns", user.Character.GetAllCooldowns())
	user.SendText(cdTxt)

	return true, nil
}
