package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/templates"

	"github.com/volte6/gomud/users"
)

func Mute(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.mute", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := users.GetByUserId(targetUserId); u != nil {

			u.Muted = true

			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-5">MUTED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	user.SendText("Could not find user.")
	return true, nil
}

func UnMute(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.mute", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := users.GetByUserId(targetUserId); u != nil {

			u.Muted = false

			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-1">UNMUTED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	user.SendText("Could not find user.")
	return true, nil
}
