package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
)

func Break(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Character.Aggro != nil {
		user.Character.Aggro = nil
		user.SendText(`You break off combat.`)
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> breaks off combat.`, user.Character.Name),
			user.UserId,
		)
	} else {
		user.SendText(`You aren't in combat!`)
	}

	return true, nil
}
