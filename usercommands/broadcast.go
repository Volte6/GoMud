package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

// Global chat room
func Broadcast(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Muted {
		user.SendText(`You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	sourceIsMod := user.Permission == users.PermissionAdmin || user.Permission == users.PermissionMod

	msg := fmt.Sprintf(`<ansi fg="black-bold">(broadcast)</ansi> <ansi fg="username">%s</ansi>: <ansi fg="yellow">%s</ansi>`, user.Character.Name, rest)

	for _, u := range users.GetAllActiveUsers() {

		if u.Deafened && !sourceIsMod {
			if u.UserId != user.UserId {
				continue
			}
		}

		u.SendText(msg)
	}

	return true, nil
}
