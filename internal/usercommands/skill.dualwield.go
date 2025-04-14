package usercommands

import (
	"errors"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/users"
)

/*
Dual WIeld
Level 1 - You can dual wield weapons that you normally couldn't. Attacks use a random weapon.
Level 2 - Occasionaly you will attack with both weapons in one round.
Level 3 - You will always attack with both weapons when Dual wielding.
Level 4 - Dual wielding incurs fewer penalties
*/
func DualWield(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.DualWield)

	if skillLevel == 0 {
		user.SendText("You haven't learned how to dual wield.")
		return true, errors.New(`you haven't learned how to dual wield`)
	}

	return Help(`dual-wield`, user, room, flags)

}
