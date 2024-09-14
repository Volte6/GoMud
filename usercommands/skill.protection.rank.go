package usercommands

import (
	"fmt"

	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
)

/*
Protection Skill
Level 2 - Front/Backrank
*/
func Rank(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Protection)

	if skillLevel < 1 {
		user.SendText("You don't know how to change your combat rank.")
		return true, fmt.Errorf("you don't know how to change your combat rank.")
	}

	party := parties.Get(userId)
	if party == nil {
		user.SendText("You must be in a party to change your combat rank.")
		return true, fmt.Errorf("you must be in a party to change your combat rank.")
	}

	if rest == `back` {
		party.SetRank(userId, `back`)
	} else if rest == `front` {
		party.SetRank(userId, `front`)
	} else {
		party.SetRank(userId, `middle`)
	}

	user.SendText(fmt.Sprintf(`You are now fighting from the <ansi fg="magenta">%s</ansi> rank.`, party.GetRank(userId)))
	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is now fighting from the <ansi fg="magenta">%s</ansi> rank.`, user.Character.Name, party.GetRank(userId)), userId)

	return true, nil
}
