package usercommands

import (
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Motd(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	user.SendText(string(configs.GetServerConfig().Motd))

	return true, nil
}
