package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/users"
)

/*
Brawling Skill
Level 1 - Enter a state of rest where health is recovered more quickly
*/
func Recover(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 1 {
		return false, nil
	}

	if user.Character.Aggro != nil {
		user.SendText("You cannot recover while in combat!")
		return true, nil
	}

	if !user.Character.TryCooldown(skills.Brawling.String(`recover`), "2 real minutes") {
		user.SendText(
			fmt.Sprintf("You need to wait %d more rounds to do that again.", user.Character.GetCooldown(skills.Brawling.String(`recover`))),
		)
		return true, nil
	}

	user.AddBuff(23, `skill`) // Warriors respite

	return true, nil
}
