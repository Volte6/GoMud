package mobcommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func LookForTrouble(rest string, mobId int) (bool, error) {

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("mob %d not found", mobId)
	}

	// Already aggroed, skip.
	if mob.Character.Aggro != nil {
		return true, nil
	}

	// Make a list of all players this gorup is hostile to in this room.
	isCharmed := mob.Character.IsCharmed()

	room := rooms.LoadRoom(mob.Character.RoomId)
	allPotentialTargets := []int{}
	nonDownedUserTargets := []int{}
	possibleMobTargets := []int{}

	//slog.Info("lookgfortrouble", "mobname", fmt.Sprintf(`%s#%d`, mob.Character.Name, mob.InstanceId))

	if !isCharmed {

		allPlayerIds := room.GetPlayers(rooms.FindAll)

		for _, playerId := range allPlayerIds {

			user := users.GetByUserId(playerId)
			if user == nil {
				continue
			}

			raceInfo := races.GetRace(user.Character.RaceId)

			// Once a player is downed mobs stop considering them a target
			// They don't see players that are sneaking...
			ignoreUser := false

			if user.Character.Health < 1 {
				ignoreUser = true
			} else if user.Character.HasBuffFlag(buffs.Hidden) {
				ignoreUser = true
			}

			entries := 1
			if party := parties.Get(playerId); party != nil {
				entries += party.ChanceToBeTargetted(playerId)
			}

			if mob.Hostile { // Does it always attack players?

				allPotentialTargets = append(allPotentialTargets, playerId)

				if !ignoreUser {
					for i := 0; i < entries; i++ {
						nonDownedUserTargets = append(nonDownedUserTargets, playerId)
					}
				}
				continue
			}

			// Does this specific mob hate this player?
			if mob.HatesRace(raceInfo.Name) {

				allPotentialTargets = append(allPotentialTargets, playerId)

				if !ignoreUser {
					for i := 0; i < entries; i++ {
						nonDownedUserTargets = append(nonDownedUserTargets, playerId)
					}
				}
				continue
			}

			// Hostility default to 5 minutes
			for _, groupName := range mob.Groups {
				// Does this group hate this player?
				if mobs.IsHostile(groupName, playerId) {

					allPotentialTargets = append(allPotentialTargets, playerId)

					if !ignoreUser {
						for i := 0; i < entries; i++ {
							nonDownedUserTargets = append(nonDownedUserTargets, playerId)
						}
					}
					break
				}
			}
		}

		// Still nothing, look for trouble in mobs they hate
		for _, considerMobInstanceId := range room.GetMobs() {
			if mobId == considerMobInstanceId {
				continue
			}

			considerMob := mobs.GetInstance(considerMobInstanceId)
			if considerMob == nil {
				continue
			}

			raceInfo := races.GetRace(mob.Character.RaceId)

			if mob.HatesMob(considerMob) || mob.HatesRace(raceInfo.Name) {
				possibleMobTargets = append(possibleMobTargets, considerMobInstanceId)
				continue
			}

			if considerMob.Character.IsCharmed() {
				for _, uid := range allPotentialTargets { // Only consider charmed mobs if they are charmed by a player this mob wants to fight
					if considerMob.Character.IsCharmed(uid) {
						possibleMobTargets = append(possibleMobTargets, considerMobInstanceId)
						break
					}
				}
			}

		}
	}

	targetUserId := 0
	targetMobInstanceId := 0

	userCt := len(nonDownedUserTargets)
	mobCt := len(possibleMobTargets)

	if userCt > 0 || mobCt > 0 {
		randRoll := util.Rand(userCt + mobCt)
		if randRoll < userCt {
			targetUserId = nonDownedUserTargets[randRoll]
		} else {
			targetMobInstanceId = possibleMobTargets[randRoll-userCt]
		}
	}

	if targetUserId > 0 {
		mob.Command(fmt.Sprintf("attack @%d", targetUserId)) // @ denotes a specific player id
	} else if targetMobInstanceId > 0 {
		mob.Command(fmt.Sprintf("attack #%d", targetMobInstanceId)) // # denotes a specific mob id
	} else {

		if mob.Despawns() {
			if mob.BoredomCounter < 255 {
				mob.BoredomCounter++
			}
		}
	}

	return true, nil
}
