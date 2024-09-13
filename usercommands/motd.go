package usercommands

import (
	"fmt"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/users"
)

func Motd(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	user.SendText(string(configs.GetConfig().Motd))

	return true, ``, nil
}
