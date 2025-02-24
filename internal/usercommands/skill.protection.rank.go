package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

/*
Protection Skill
Level 2 - Front/Backrank
*/
func Rank(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Protection)

	if skillLevel < 1 {
		user.SendText("You don't know how to change your combat rank.")
		return true, fmt.Errorf("you don't know how to change your combat rank.")
	}

	party := parties.Get(user.UserId)
	if party == nil {
		user.SendText("You must be in a party to change your combat rank.")
		return true, fmt.Errorf("you must be in a party to change your combat rank.")
	}

	if rest == `back` {
		party.SetRank(user.UserId, `back`)
	} else if rest == `front` {
		party.SetRank(user.UserId, `front`)
	} else {
		party.SetRank(user.UserId, `middle`)
	}

	user.SendText(fmt.Sprintf(`You are now fighting from the <ansi fg="magenta">%s</ansi> rank.`, party.GetRank(user.UserId)))
	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is now fighting from the <ansi fg="magenta">%s</ansi> rank.`, user.Character.Name, party.GetRank(user.UserId)), user.UserId)

	return true, nil
}
