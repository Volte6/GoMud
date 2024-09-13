package usercommands

import (
	"fmt"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Flee(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Aggro == nil {
		user.SendText(`You aren't in combat!`)
	} else {
		user.SendText(`You attempt to flee...`)
		user.Character.Aggro.Type = characters.Flee
	}

	response.Handled = true
	return response, nil
}
