package usercommands

import (
	"fmt"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Cooldowns(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf(`user %d not found`, userId)
	}

	cdTxt, _ := templates.Process("character/cooldowns", user.Character.GetAllCooldowns())
	user.SendText(cdTxt)

	return true, nil
}
