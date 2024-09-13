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
	"github.com/volte6/mud/util"
)

func Aid(rest string, mobId int) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	// Load current room details
	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	raceInfo := races.GetRace(mob.Character.RaceId)
	if !raceInfo.KnowsFirstAid {

		mob.Command(`emote doesn't know first aid.`)

		response.Handled = true
		return response, nil
	}

	if !room.IsCalm() {
		response.Handled = true
		return response, nil
	}

	if rest == `` {
		response.Handled = true
		return response, nil
	}

	aidPlayerId, _ := room.FindByName(rest, rooms.FindDowned)

	if aidPlayerId > 0 {

		p := users.GetByUserId(aidPlayerId)

		if p != nil {

			if p.Character.Health > 0 {
				response.Handled = true
				return response, nil
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
			if res, err := scripting.TrySpellScriptEvent(`onCast`, 0, mobId, spellAggro); err == nil {
				continueCasting = res.Handled
			}

			if continueCasting {

				spellInfo := spells.GetSpell(`aidskill`)
				mob.Character.CancelBuffsWithFlag(buffs.Hidden)
				mob.Character.SetCast(spellInfo.WaitRounds, spellAggro)
			}

		}

	}

	response.Handled = true
	return response, nil
}
