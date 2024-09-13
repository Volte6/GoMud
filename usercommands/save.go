package usercommands

import (
	"fmt"

	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Save(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	response.SendUserMessage(userId, "Saving...")
	users.SaveUser(*user)
	response.SendUserMessage(userId, "done.")

	response.Handled = true
	return response, nil
}
