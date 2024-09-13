package usercommands

import (
	"fmt"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Motd(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	user.SendText(string(configs.GetConfig().Motd))

	response.Handled = true
	return response, nil
}
