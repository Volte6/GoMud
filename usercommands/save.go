package usercommands

import (
	"fmt"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Save(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	response.SendUserMessage(userId, "Saving...", true)
	users.SaveUser(*user)
	response.SendUserMessage(userId, "done.", true)

	response.Handled = true
	return response, nil
}
