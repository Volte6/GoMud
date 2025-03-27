package usercommands

import (
	"strings"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

/*
* Role Permissions:
* syslogs 				(All)
 */
func SysLogs(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.syslogs", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	args := util.SplitButRespectQuotes(rest)

	if args[0] == "off" {

		events.AddToQueue(events.Log{FollowRemove: user.ConnectionId()})
		user.SendText(`Log follow disabled.`)

		return true, nil
	}

	events.AddToQueue(events.Log{FollowAdd: user.ConnectionId(), Level: strings.ToUpper(args[0])})
	user.SendText(`Log follow enabled. Use <ansi fg="command">syslogs off</ansi> to turn it off.`)

	return true, nil
}
