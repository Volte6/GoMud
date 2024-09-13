package usercommands

import (
	"fmt"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Break(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Aggro != nil {
		user.Character.Aggro = nil
		response.SendUserMessage(userId, `You break off combat.`, true)
		response.SendRoomMessage(user.Character.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> breaks off combat.`, user.Character.Name),
			true)
	} else {
		response.SendUserMessage(userId, `You aren't in combat!`, true)
	}

	response.Handled = true
	return response, nil
}
