package usercommands

import (
	"fmt"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/users"
)

func Flee(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Aggro == nil {
		user.SendText(`You aren't in combat!`)
	} else {
		user.SendText(`You attempt to flee...`)
		user.Character.Aggro.Type = characters.Flee
	}

	return true, ``, nil
}
