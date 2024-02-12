package mobcommands

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Suicide(rest string, mobId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewMobCommandResponse(mobId)

	// Load user details
	mob := mobs.GetInstance(mobId)
	if mob == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("mob %d not found", mobId)
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, mob.Character.RoomId)
	}

	slog.Info(`Mob Death`, `name`, mob.Character.Name)

	// Send a death msg to everyone in the room.
	response.SendRoomMessage(mob.Character.RoomId,
		fmt.Sprintf(`<ansi fg="mobname">%s</ansi> has died.`, mob.Character.Name),
		true)

	mobXP := mob.Character.XPTL(mob.Character.Level - 1)

	xpVal := mobXP / 150

	xpVariation := xpVal / 100
	if xpVariation < 1 {
		xpVariation = 1
	}

	partyTracker := map[int]int{} // key is party leader ID, value is how much will be shared.

	if len(mob.DamageTaken) > 0 {

		xpVal = xpVal / len(mob.DamageTaken)        // Div by number of players that beat him up
		xpVal += ((util.Rand(3) - 1) * xpVariation) // a little bit of variation

		totalPlayerLevels := 0
		for uId, _ := range mob.DamageTaken {
			if user := users.GetByUserId(uId); user != nil {
				totalPlayerLevels += user.Character.Level
			}
		}

		attackerCt := len(mob.DamageTaken)

		xpMsg := `You gained <ansi fg="experience">%d experience points</ansi>%s!`
		for uId, _ := range mob.DamageTaken {
			if user := users.GetByUserId(uId); user != nil {

				if res, err := scripting.TryMobScriptEvent(`onDie`, mob.InstanceId, uId, `user`, map[string]any{`attackerCount`: attackerCt}, cmdQueue); err == nil {
					response.AbsorbMessages(res)
				}

				p := parties.Get(user.UserId)

				// Not in a party? Great give them the xp.
				if p == nil {

					xpScaler := float64(mob.Character.Level) / float64(totalPlayerLevels)
					//if xpScaler > 1 {
					xpVal = int(math.Ceil(float64(xpVal) * xpScaler))
					//}

					grantXP, xpScale := user.Character.GrantXP(xpVal)

					xpMsgExtra := ``
					if xpScale != 100 {
						xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
					}

					response.SendUserMessage(user.UserId,
						fmt.Sprintf(xpMsg, grantXP, xpMsgExtra),
						true)

					continue
				}

				if _, ok := partyTracker[p.LeaderUserId]; !ok {
					partyTracker[p.LeaderUserId] = 0
				}
				partyTracker[p.LeaderUserId] += xpVal

			}
		}

	}

	if len(partyTracker) > 0 {

		xpMsg := `You gained <ansi fg="yellow-bold">%d experience points</ansi>%s!`
		for leaderId, xp := range partyTracker {
			if p := parties.Get(leaderId); p != nil {

				allMembers := p.GetMembers()
				xpSplit := xp / len(allMembers)

				slog.Info(`Party XP`, `totalXP`, xp, `splitXP`, xpSplit, `memberCt`, len(allMembers))

				for _, memberId := range allMembers {

					if user := users.GetByUserId(memberId); user != nil {

						grantXP, xpScale := user.Character.GrantXP(xpSplit)

						xpMsgExtra := ``
						if xpScale != 100 {
							xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
						}

						response.SendUserMessage(user.UserId,
							fmt.Sprintf(xpMsg, grantXP, xpMsgExtra),
							true)
					}

				}

			}
		}

	}

	// Check for any dropped loot...
	for _, item := range mob.Character.Items {
		msg := fmt.Sprintf(`<ansi fg="item">%s</ansi> drops to the ground.`, item.Name())
		response.SendRoomMessage(mob.Character.RoomId, msg, true)
		room.AddItem(item, false)
	}

	allWornItems := mob.Character.Equipment.GetAllItems()

	for _, item := range allWornItems {

		roll := util.Rand(100)

		util.LogRoll(`Drop Item`, roll, mob.ItemDropChance)

		if roll >= mob.ItemDropChance {
			continue
		}

		msg := fmt.Sprintf(`<ansi fg="item">%s</ansi> drops to the ground.`, item.Name())
		response.SendRoomMessage(mob.Character.RoomId, msg, true)
		room.AddItem(item, false)
	}

	if mob.Character.Gold > 0 {
		msg := fmt.Sprintf(`<ansi fg="yellow-bold">%d gold</ansi> drops to the ground.`, mob.Character.Gold)
		response.SendRoomMessage(mob.Character.RoomId, msg, true)
		room.Gold += mob.Character.Gold
	}

	// Destroy any record of this mob.
	mobs.DestroyInstance(mob.InstanceId)

	// Clean up mob from room...
	if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
		r.CleanupMobSpawns(false)
	}

	// Remove from current room
	room.RemoveMob(mob.InstanceId)

	response.Handled = true
	return response, nil
}
