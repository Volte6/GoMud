package mobcommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Give(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		return true, nil
	}

	var giveWho string = args[len(args)-1]
	args = args[:len(args)-1]
	var giveWhat string = strings.Join(args, " ")

	var giveItem items.Item = items.Item{}
	var giveGoldAmount int = 0

	if len(giveWhat) > 4 && giveWhat[len(giveWhat)-4:] == "gold" {

		g, _ := strconv.ParseInt(giveWhat[0:len(giveWhat)-5], 10, 32)
		giveGoldAmount = int(g)

		if giveGoldAmount > mob.Character.Gold {
			return true, nil
		}

	} else {

		var found bool = false

		// Check whether the user has an item in their inventory that matches
		giveItem, found = mob.Character.FindInBackpack(giveWhat)

		if !found {
			return true, nil
		}

	}

	playerId, mobId := room.FindByName(giveWho)

	if playerId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if giveItem.ItemId > 0 {
			targetUser.Character.StoreItem(giveItem)
			mob.Character.RemoveItem(giveItem)

			events.AddToQueue(events.ItemOwnership{
				MobInstanceId: mob.InstanceId,
				Item:          giveItem,
				Gained:        false,
			})

			events.AddToQueue(events.ItemOwnership{
				UserId: targetUser.UserId,
				Item:   giveItem,
				Gained: true,
			})

			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gives you their <ansi fg="item">%s</ansi>.`, mob.Character.Name, giveItem.DisplayName()),
			)

		} else if giveGoldAmount > 0 {

			targetUser.Character.Gold += giveGoldAmount
			mob.Character.Gold -= giveGoldAmount

			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gives you <ansi fg="gold">%d gold</ansi>.`, mob.Character.Name, giveGoldAmount),
			)

		}

		return true, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		mob.Character.CancelBuffsWithFlag(buffs.Hidden)

		m := mobs.GetInstance(mobId)

		if m != nil {

			// Swap the item location
			if giveItem.ItemId > 0 {
				m.Character.StoreItem(giveItem)
				mob.Character.RemoveItem(giveItem)

				events.AddToQueue(events.ItemOwnership{
					MobInstanceId: mob.InstanceId,
					Item:          giveItem,
					Gained:        false,
				})

				events.AddToQueue(events.ItemOwnership{
					MobInstanceId: m.InstanceId,
					Item:          giveItem,
					Gained:        true,
				})

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gave their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, giveItem.DisplayName(), m.Character.Name),
				)
			} else if giveGoldAmount > 0 {

				m.Character.Gold += giveGoldAmount
				mob.Character.Gold -= giveGoldAmount

				room.SendText(
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> gave some gold to <ansi fg="mobname">%s</ansi>.`, mob.Character.Name, m.Character.Name),
				)
			}

		}

	}

	return true, nil
}
