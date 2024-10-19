package mobcommands

import (
	"github.com/volte6/gomud/buffs"
	"github.com/volte6/gomud/characters"
	"github.com/volte6/gomud/mobs"
	"github.com/volte6/gomud/races"
	"github.com/volte6/gomud/rooms"
	"github.com/volte6/gomud/scripting"
	"github.com/volte6/gomud/spells"
	"github.com/volte6/gomud/users"
)

func Aid(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

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
			if allowToContinue, err := scripting.TrySpellScriptEvent(`onCast`, 0, mob.InstanceId, spellAggro); err == nil {
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
