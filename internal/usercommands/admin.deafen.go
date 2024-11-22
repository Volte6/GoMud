package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"

	"github.com/volte6/gomud/internal/users"
)

func Deafen(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.deafen", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := users.GetByUserId(targetUserId); u != nil {

			u.Deafened = true

			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-5">DEAFENED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	user.SendText("Could not find user.")
	return true, nil
}

func UnDeafen(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.deafen", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := users.GetByUserId(targetUserId); u != nil {

			u.Deafened = false

			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-1">UNDEAFENED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	user.SendText("Could not find user.")
	return true, nil
}
