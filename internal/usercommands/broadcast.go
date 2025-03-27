package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

// Global chat room
func Broadcast(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user.Muted {
		user.SendText(`You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	msg := fmt.Sprintf(`<ansi fg="broadcast-prefix">(broadcast)</ansi> <ansi fg="username">%s</ansi>: <ansi fg="broadcast-body">%s</ansi>`, user.Character.Name, rest)

	events.AddToQueue(events.Broadcast{
		Text:            msg + term.CRLFStr,
		IsCommunication: true,
		SourceIsMod:     user.Role != users.RoleUser,
	})

	return true, nil
}
