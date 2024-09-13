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

	user.SendText("Saving...")
	users.SaveUser(*user)
	user.SendText("done.")

	response.Handled = true
	return response, nil
}
