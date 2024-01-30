package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Aid(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Protection)

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to provide aid.", true)
		response.Handled = true
		return response, fmt.Errorf("you don't know how to provide aid")
	}

	if skillLevel < 3 && !room.IsCalm() {
		response.SendUserMessage(userId, "You can only do that in calm rooms!", true)
		response.Handled = true
		return response, nil
	}

	aidPlayerId, _ := room.FindByName(rest, rooms.FindDowned)

	if aidPlayerId == userId {
		aidPlayerId = 0
	}

	if aidPlayerId > 0 {

		p := users.GetByUserId(aidPlayerId)

		if p != nil {

			if p.Character.Health > 0 {
				response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is not in need of aid!`, p.Character.Name), true)
				response.Handled = true
				return response, nil
			}

			if user.Character.Aggro != nil {
				response.SendUserMessage(userId, "You are too busy to aid anyone!", true)
				response.Handled = true
				return response, nil
			}

			user.Character.CancelBuffsWithFlag(buffs.Hidden)

			user.Character.SetAid(p.UserId, 2)

			response.SendUserMessage(user.UserId, fmt.Sprintf(`You prepare to provide aid to <ansi fg="username">%s</ansi>.`, p.Character.Name), true)
			response.SendUserMessage(p.UserId, fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to apply first aid on you.`, user.Character.Name), true)
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to provide aid to <ansi fg="username">%s</ansi>.`, user.Character.Name, p.Character.Name), true, user.UserId, p.UserId)
		}

		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId, "Aid whom?", true)
	response.Handled = true
	return response, nil
}
