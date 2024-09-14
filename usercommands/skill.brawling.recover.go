package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
)

/*
Brawling Skill
Level 1 - Enter a state of rest where health is recovered more quickly
*/
func Recover(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 1 {
		return false, nil
	}

	if user.Character.Aggro != nil {
		user.SendText("You cannot recover while in combat!")
		return true, nil
	}

	if !user.Character.TryCooldown(skills.Brawling.String(`recover`), 25) {
		user.SendText(
			fmt.Sprintf("You need to wait %d more rounds to do that again.", user.Character.GetCooldown(skills.Brawling.String(`recover`))),
		)
		return true, nil
	}

	applyBuffId := 23
	if skillLevel >= 3 {
		applyBuffId = 25
	} else if skillLevel >= 2 {
		applyBuffId = 24
	}

	events.AddToQueue(events.Buff{
		UserId:        userId,
		MobInstanceId: 0,
		BuffId:        applyBuffId,
	})

	return true, nil
}
