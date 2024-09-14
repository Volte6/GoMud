package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/users"
)

func Aid(rest string, mobId int) (bool, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	raceInfo := races.GetRace(mob.Character.RaceId)
	if !raceInfo.KnowsFirstAid {

		mob.Command(`emote doesn't know first aid.`)

		return true, nil
	}

	if !room.IsCalm() {
		return true, nil
	}

	if rest == `` {
		return true, nil
	}

	aidPlayerId, _ := room.FindByName(rest, rooms.FindDowned)

	if aidPlayerId > 0 {

		p := users.GetByUserId(aidPlayerId)

		if p != nil {

			if p.Character.Health > 0 {
				return true, nil
			}

			mob.Character.CancelBuffsWithFlag(buffs.Hidden)

			// Set spell Aid
			spellAggro := characters.SpellAggroInfo{
				SpellId:              `aidskill`,
				SpellRest:            ``,
				TargetUserIds:        []int{aidPlayerId},
				TargetMobInstanceIds: []int{},
			}

			continueCasting := true
			if allowToContinue, err := scripting.TrySpellScriptEvent(`onCast`, 0, mobId, spellAggro); err == nil {
				continueCasting = allowToContinue
			}

			if continueCasting {

				spellInfo := spells.GetSpell(`aidskill`)
				mob.Character.CancelBuffsWithFlag(buffs.Hidden)
				mob.Character.SetCast(spellInfo.WaitRounds, spellAggro)
			}

		}

	}

	return true, nil
}
