package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

/*
Brawling Skill
Level 1 - Enter a state of rest where health is recovered more quickly
*/
func Recover(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

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

	events.AddToQueue(events.Buff{
		UserId:        user.UserId,
		MobInstanceId: 0,
		BuffId:        23, // Warriors respite
	})

	return true, nil
}
