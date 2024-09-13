package usercommands

import (
	"fmt"

	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Protection Skill
Level 2 - Front/Backrank
*/
func Rank(rest string, userId int) (util.MessageQueue, error) {

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

	if skillLevel < 1 {
		response.SendUserMessage(userId, "You don't know how to change your combat rank.")
		response.Handled = true
		return response, fmt.Errorf("you don't know how to change your combat rank.")
	}

	party := parties.Get(userId)
	if party == nil {
		response.SendUserMessage(userId, "You must be in a party to change your combat rank.")
		response.Handled = true
		return response, fmt.Errorf("you must be in a party to change your combat rank.")
	}

	if rest == `back` {
		party.SetRank(userId, `back`)
	} else if rest == `front` {
		party.SetRank(userId, `front`)
	} else {
		party.SetRank(userId, `middle`)
	}

	response.SendUserMessage(userId, fmt.Sprintf(`You are now fighting from the <ansi fg="magenta">%s</ansi> rank.`, party.GetRank(userId)))
	response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now fighting from the <ansi fg="magenta">%s</ansi> rank.`, user.Character.Name, party.GetRank(userId)), userId)

	response.Handled = true
	return response, nil
}
