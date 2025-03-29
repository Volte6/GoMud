package usercommands

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Exits(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	details := rooms.GetDetails(room, user)

	exitTxt, _ := templates.Process("descriptions/exits", details, user.UserId)
	user.SendText(exitTxt)

	return true, nil
}
