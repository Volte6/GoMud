package usercommands

import (
	"fmt"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

func Break(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if user.Character.Aggro != nil {
		user.Character.Aggro = nil
		user.SendText(`You break off combat.`)
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> breaks off combat.`, user.Character.Name),
			userId,
		)
	} else {
		user.SendText(`You aren't in combat!`)
	}

	return true, nil
}
