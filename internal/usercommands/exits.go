package usercommands

import (
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Exits(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	details := rooms.GetDetails(room, user)

	exitTxt, _ := templates.Process("descriptions/exits", details)
	user.SendText(exitTxt)

	return true, nil
}
