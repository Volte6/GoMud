package usercommands

import (
	"github.com/volte6/gomud/configs"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

func Motd(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	user.SendText(string(configs.GetConfig().Motd))

	return true, nil
}
