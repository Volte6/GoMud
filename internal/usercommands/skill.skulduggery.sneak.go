package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/users"
)

/*
SkullDuggery Skill
Level 1 - Sneak
*/
func Sneak(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Skulduggery)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 1 {
		return false, nil
	}

	// Must be sneaking
	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	if isSneaking {
		user.SendText("You're already hidden!")
		return true, nil
	}

	if user.Character.Aggro != nil {
		user.SendText("You can't do that while in combat!")
		return true, nil
	}

	if room := rooms.LoadRoom(user.Character.RoomId); room != nil && !room.IsCalm() {
		user.SendText("You can only do that in calm rooms!")
		return true, nil
	}

	user.AddBuff(9, `skill`)

	return true, nil
}
