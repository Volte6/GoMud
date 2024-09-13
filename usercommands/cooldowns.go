package usercommands

import (
	"fmt"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Cooldowns(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
	}

	cdTxt, _ := templates.Process("character/cooldowns", user.Character.GetAllCooldowns())
	response.SendUserMessage(userId, cdTxt)

	response.Handled = true
	return response, nil
}
