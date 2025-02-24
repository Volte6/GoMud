package usercommands

import (
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

/*
SkullDuggery Skill
Level 1 - Sneak
*/
func Sneak(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

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

	events.AddToQueue(events.Buff{
		UserId:        user.UserId,
		MobInstanceId: 0,
		BuffId:        9,
	})

	return true, nil
}
