package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
SkullDuggery Skill
Level 1 - Sneak
*/
func Sneak(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Skulduggery)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 1 {
		return response, nil
	}

	// Must be sneaking
	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)
	if isSneaking {
		user.SendText("You're already hidden!")
		response.Handled = true
		return response, nil
	}

	if user.Character.Aggro != nil {
		user.SendText("You can't do that while in combat!")
		response.Handled = true
		return response, nil
	}

	if room := rooms.LoadRoom(user.Character.RoomId); room != nil && !room.IsCalm() {
		user.SendText("You can only do that in calm rooms!")
		response.Handled = true
		return response, nil
	}

	events.AddToQueue(events.Buff{
		UserId:        userId,
		MobInstanceId: 0,
		BuffId:        9,
	})

	response.Handled = true
	return response, nil
}
