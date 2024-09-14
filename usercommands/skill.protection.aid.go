package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/users"
)

/*
Protection Skill
Level 1 - Aid (revive) a player
Level 3 - Aid (revive) a player, even during combat
*/
func Aid(rest string, userId int) (bool, error) {

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

	if skillLevel == 0 {
		user.SendText("You don't know how to provide aid.")
		return true, fmt.Errorf("you don't know how to provide aid")
	}

	if skillLevel < 3 && !room.IsCalm() {
		user.SendText("You can only do that in calm rooms!")
		return true, nil
	}

	aidPlayerId, _ := room.FindByName(rest, rooms.FindDowned)

	if aidPlayerId == userId {
		aidPlayerId = 0
	}

	if aidPlayerId > 0 {

		p := users.GetByUserId(aidPlayerId)

		if p != nil {

			if p.Character.Health > 0 {
				user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is not in need of aid!`, p.Character.Name))
				return true, nil
			}

			if user.Character.Aggro != nil {
				user.SendText("You are too busy to aid anyone!")
				return true, nil
			}

			// Set spell Aid
			spellAggro := characters.SpellAggroInfo{
				SpellId:              `aidskill`,
				SpellRest:            ``,
				TargetUserIds:        []int{aidPlayerId},
				TargetMobInstanceIds: []int{},
			}

			continueCasting := true
			if handled, err := scripting.TrySpellScriptEvent(`onCast`, userId, 0, spellAggro); err == nil {
				continueCasting = handled
			}

			if continueCasting {
				spellInfo := spells.GetSpell(`aidskill`)
				user.Character.CancelBuffsWithFlag(buffs.Hidden)
				user.Character.SetCast(spellInfo.WaitRounds, spellAggro)
			}

		}

		return true, nil
	}

	user.SendText("Aid whom?")
	return true, nil
}
