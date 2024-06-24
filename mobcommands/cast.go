package mobcommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Cast(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		response.Handled = true
		return response, nil
	}

	spellName := args[0]
	spellArg := strings.Join(args[1:], ` `)

	spellInfo := spells.GetSpell(spellName)

	if spellInfo == nil {
		response.Handled = true
		return response, nil
	}

	if mob.Character.Mana < spellInfo.Cost {
		response.Handled = true
		return response, nil
	}

	targetPlayerId := 0
	targetMobInstanceId := 0

	if spellArg != `` {
		targetPlayerId, targetMobInstanceId = room.FindByName(spellArg)
	}

	spellAggro := characters.SpellAggroInfo{
		SpellId:              spellInfo.SpellId,
		SpellRest:            ``,
		TargetUserIds:        make([]int, 0),
		TargetMobInstanceIds: make([]int, 0),
	}

	if spellInfo.Type == spells.Neutral {

		spellAggro.SpellRest = spellArg

	} else if spellInfo.Type == spells.HelpSingle {

		if spellArg == `` {

			// No target specified? Default to self
			spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mobId)

		} else {

			if targetPlayerId > 0 {
				spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, targetPlayerId)
			} else if targetMobInstanceId > 0 {
				spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, targetMobInstanceId)
			}

		}

	} else if spellInfo.Type == spells.HarmSingle {

		if spellArg == `` {

			if mob.Character.Aggro != nil {
				// No target specified? Default to self
				if mob.Character.Aggro.UserId > 0 {
					spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, mob.Character.Aggro.UserId)
				} else if mob.Character.Aggro.MobInstanceId > 0 {
					spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mob.Character.Aggro.MobInstanceId)
				}
			} else {

				playersFightingMobs := room.GetPlayers(rooms.FindFightingMob)
				if len(playersFightingMobs) > 0 {

					for _, pUserId := range playersFightingMobs {

						if u := users.GetByUserId(pUserId); u != nil {
							if u.Character.IsAggro(0, mobId) {
								spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, pUserId)
								break
							}
						}

					}

				}

				// If no mobs found, try finding an aggro player
				if len(spellAggro.TargetMobInstanceIds) < 1 {
					mobsFightingMobs := room.GetMobs(rooms.FindFightingMob)
					if len(mobsFightingMobs) > 0 {

						for _, mId := range mobsFightingMobs {

							if m := mobs.GetInstance(mId); m != nil {
								if m.Character.IsAggro(0, mobId) {
									spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mId)
									break
								}
							}

						}

					}
				}

			}

		} else {

			if targetPlayerId > 0 {
				spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, targetPlayerId)
			} else if targetMobInstanceId > 0 {
				spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, targetMobInstanceId)
			}

		}

	} else if spellInfo.Type == spells.HelpMulti {

		spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mobId)

		if !mob.Character.IsCharmed() {

			for _, mobInstId := range room.GetMobs() {

				if m := mobs.GetInstance(mobInstId); m != nil {
					// Cast on same kind
					if m.MobId == mob.MobId {
						spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mobInstId)
					}
				}
			}

		}

	} else if spellInfo.Type == spells.HarmMulti {

		// Targets all mobs aggro towards player
		// Targets all players aggro towards player and their parties

		// If not currently aggro, only targets all mobs in the room

		mobsFightingMobs := room.GetMobs(rooms.FindFightingMob)
		for _, mobInstId := range mobsFightingMobs {
			if m := mobs.GetInstance(mobInstId); m != nil {
				if m.Character.IsAggro(0, mobId) || m.HatesRace(m.Character.Race()) {
					spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mobInstId)
				}
			}
		}

		playersFightingMobs := room.GetPlayers(rooms.FindFightingMob)
		for _, uId := range playersFightingMobs {
			if u := users.GetByUserId(uId); u != nil {
				if u.Character.IsAggro(0, mobId) {
					spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, uId)
				}
			}
		}

	}

	if len(spellAggro.TargetUserIds) > 0 || len(spellAggro.TargetMobInstanceIds) > 0 || len(spellAggro.SpellRest) > 0 {

		continueCasting := true
		if res, err := scripting.TrySpellScriptEvent(`onCast`, 0, mobId, spellAggro, cmdQueue); err == nil {
			response.AbsorbMessages(res)
			continueCasting = res.Handled
		}

		if continueCasting {
			mob.Character.Mana -= spellInfo.Cost
			mob.Character.SetCast(spellInfo.WaitRounds, spellAggro)
		}

	}

	response.Handled = true
	return response, nil
}
