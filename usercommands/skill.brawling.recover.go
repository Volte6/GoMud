package usercommands

import (
	"fmt"

	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Recover(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 1 {
		return response, nil
	}

	if user.Character.Aggro != nil {
		response.SendUserMessage(userId, "You cannot recover while in combat!", true)
		response.Handled = true
		return response, nil
	}

	if !user.Character.TryCooldown(skills.Brawling.String(`recover`), 25) {
		response.SendUserMessage(userId,
			fmt.Sprintf("You need to wait %d more rounds to do that again.", user.Character.GetCooldown(skills.Brawling.String(`recover`))),
			true)
		response.Handled = true
		return response, nil
	}

	if skillLevel >= 3 {
		cmdQueue.QueueBuff(userId, 0, 25)
	} else if skillLevel >= 2 {
		cmdQueue.QueueBuff(userId, 0, 24)
	} else {
		cmdQueue.QueueBuff(userId, 0, 23)
	}

	response.Handled = true
	return response, nil
}
