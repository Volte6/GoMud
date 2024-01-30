package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func DualWield(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.DualWield)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You haven't learned how to dual wield.", true)
		response.Handled = true
		return response, errors.New(`you haven't learned how to dual wield`)
	}

	return Help(`dual-wield`, userId, cmdQueue)

}
