package usercommands

import (
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Motd(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	user.SendText(string(configs.GetConfig().Motd))

	return true, nil
}
