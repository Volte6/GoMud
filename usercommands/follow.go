package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/users"
)

func Follow(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest == "" {
		user.SendText("Follow whom?")
		return true, nil
	}

	playerId, _ := room.FindByName(rest)

	if playerId == 0 {
		user.SendText(fmt.Sprintf(`%s - not found`, rest))
		return true, nil
	}

	if playerId == user.UserId {
		user.SendText(`You can't follow yourself`)
		return true, nil
	}

	if playerId > 0 {

		followUser := users.GetByUserId(playerId)

		user.SendText(
			fmt.Sprintf(`You follow <ansi fg="username">%s</ansi>.`, followUser.Character.Name),
		)

		followUser.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> is following you.`, user.Character.Name),
		)

		followUser.Character.AddFollower(user.UserId)
	}

	return true, nil
}
