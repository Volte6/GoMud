package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
)

/*
Dual WIeld
Level 1 - You can dual wield weapons that you normally couldn't. Attacks use a random weapon.
Level 2 - Occasionaly you will attack with both weapons in one round.
Level 3 - You will always attack with both weapons when Dual wielding.
Level 4 - Dual wielding incurs fewer penalties
*/
func DualWield(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.DualWield)

	if skillLevel == 0 {
		user.SendText("You haven't learned how to dual wield.")
		return true, errors.New(`you haven't learned how to dual wield`)
	}

	return Help(`dual-wield`, userId)

}
