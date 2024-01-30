package usercommands

import (
	"fmt"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Broadcast(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	response.SendRoomMessage(0, fmt.Sprintf(`<ansi fg="black" bold="true'>(broadcast)</ansi> <ansi fg="username">%s</ansi>: <ansi fg="yellow">%s</ansi>`, user.Character.Name, rest), true)

	response.Handled = true
	return response, nil
}
