package usercommands

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Who(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	details := rooms.GetDetails(room, user)

	whoTxt, _ := templates.Process("descriptions/who", details)
	user.SendText(whoTxt)

	return true, nil
}
