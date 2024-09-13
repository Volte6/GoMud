package usercommands

import (
	"fmt"

	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Follow(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if rest == "" {
		user.SendText("Follow whom?")
		response.Handled = true
		return response, nil
	}

	playerId, _ := room.FindByName(rest)
	if playerId == userId {
		playerId = 0
	}

	if playerId > 0 {

		followUser := users.GetByUserId(playerId)

		user.SendText(
			fmt.Sprintf(`You follow <ansi fg="username">%s</ansi>.`, followUser.Character.Name),
		)

		followUser.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> is following you.`, user.Character.Name),
		)

		followUser.Character.AddFollower(userId)

	}

	response.Handled = true
	return response, nil
}
