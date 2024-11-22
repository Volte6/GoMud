package usercommands

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Motd(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	user.SendText(string(configs.GetConfig().Motd))

	return true, nil
}
