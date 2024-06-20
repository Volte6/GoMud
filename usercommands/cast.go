package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Cast(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		response.SendUserMessage(userId, "Cast What? At Whom?", true)
		response.Handled = true
		return response, nil
	}

	spellName := args[0]
	spellArg := strings.Join(args[1:], ` `)

	spellInfo := spells.GetSpell(spellName)

	if spellInfo == nil || !user.Character.HasSpell(spellName) {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't know a spell called <ansi fg="spellname">%s</ansi>.`, spellName), true)
		response.Handled = true
		return response, nil
	}

	if user.Character.Mana < spellInfo.Cost {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't have enough mana to cast <ansi fg="spellname">%s</ansi>.`, spellName), true)
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
			spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, userId)

		} else {

			if targetPlayerId > 0 {
				spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, userId)
			} else if targetMobInstanceId > 0 {
				spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, targetMobInstanceId)
			}

		}

	} else if spellInfo.Type == spells.HarmSingle {

		if spellArg == `` {

			if user.Character.Aggro != nil {
				// No target specified? Default to self
				if user.Character.Aggro.UserId > 0 {
					spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, userId)
				} else if user.Character.Aggro.MobInstanceId > 0 {
					spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, user.Character.Aggro.MobInstanceId)
				}
			} else {

				fightingMobs := room.GetMobs(rooms.FindFightingPlayer)
				if len(fightingMobs) > 0 {

					for _, mobInstId := range fightingMobs {

						if mob := mobs.GetInstance(mobInstId); mob != nil && mob.Character.IsAggro(userId, 0) {
							spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mobInstId)
							break
						}

					}

				}

				// If no mobs found, try finding an aggro player
				if len(spellAggro.TargetMobInstanceIds) < 1 {
					fightingPlayers := room.GetPlayers(rooms.FindFightingPlayer)
					if len(fightingPlayers) > 0 {

						for _, fUserId := range fightingPlayers {

							if u := users.GetByUserId(fUserId); u != nil && u.Character.IsAggro(userId, 0) {
								spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, fUserId)
								break
							}

						}

					}
				}

			}

		} else {

			if targetPlayerId > 0 {
				spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, userId)
			} else if targetMobInstanceId > 0 {
				spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, targetMobInstanceId)
			}

		}

	} else if spellInfo.Type == spells.HelpMulti {

		spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, userId)

		// Targets self and all in party
		if p := parties.Get(userId); p != nil {

			for _, partyUserId := range p.GetMembers() {

				if partyUserId == userId {
					continue
				}

				spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, partyUserId)

			}

			for _, partyMobId := range p.GetMobs() {

				spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, partyMobId)

			}

		}

	} else if spellInfo.Type == spells.HarmMulti {

		// Targets all mobs aggro towards player
		// Targets all players aggro towards player and their parties

		// If not currently aggro, only targets all mobs in the room

		fightingMobs := room.GetMobs(rooms.FindFightingPlayer)
		for _, mobInstId := range fightingMobs {
			if m := mobs.GetInstance(mobInstId); m != nil && m.Character.IsAggro(userId, 0) {
				spellAggro.TargetMobInstanceIds = append(spellAggro.TargetMobInstanceIds, mobInstId)
			}
		}

		fightingPlayers := room.GetPlayers(rooms.FindFightingPlayer)
		for _, uId := range fightingPlayers {
			if u := users.GetByUserId(uId); u != nil && u.Character.IsAggro(userId, 0) {
				spellAggro.TargetUserIds = append(spellAggro.TargetUserIds, uId)
			}
		}

		if len(spellAggro.TargetUserIds) < 1 && len(spellAggro.TargetMobInstanceIds) < 1 {
			// No targets found, default to all mobs in the room
			spellAggro.TargetMobInstanceIds = fightingMobs
		}

	}

	if len(spellAggro.TargetUserIds) > 0 || len(spellAggro.TargetMobInstanceIds) > 0 || len(spellAggro.SpellRest) > 0 {

		continueCasting := true
		if res, err := scripting.TrySpellScriptEvent(`onCast`, userId, 0, spellAggro, cmdQueue); err == nil {
			response.AbsorbMessages(res)
			continueCasting = res.Handled
		}

		if continueCasting {
			user.Character.Mana -= spellInfo.Cost
			user.Character.SetCast(spellInfo.WaitRounds, spellAggro)
		}

	} else {

		response.SendUserMessage(userId, `Couldn't find a target for the spell.`, true)

	}

	response.Handled = true
	return response, nil
}
