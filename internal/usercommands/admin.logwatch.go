package usercommands

import (
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func LogWatch(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.logwatch", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	if rest == "on" {
		events.AddToQueue(events.Log{FollowAdd: user.ConnectionId()})
		user.SendText(`Log follow enabled. Use <ansi fg="command">logwatch off</ansi> to turn it off.`)
	} else if rest == "off" {
		events.AddToQueue(events.Log{FollowRemove: user.ConnectionId()})
		user.SendText(`Log follow disabled. Use <ansi fg="command">logwatch on</ansi> to turn it on.`)
	}

	return true, nil
}
