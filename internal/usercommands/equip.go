package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/scripting"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Equip(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "all" {
		return Gearup(``, user, room, flags)
	}

	if rest == "" {
		user.SendText(`Wear WHAT?`)
		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to wear.`, rest))
	} else {

		iSpec := matchItem.GetSpec()
		if iSpec.Type != items.Weapon && iSpec.Subtype != items.Wearable {
			user.SendText(
				fmt.Sprintf(`Your <ansi fg="item">%s</ansi> doesn't look very fashionable.`, matchItem.DisplayName()),
			)
			return true, nil
		}

		// Swap the item location
		oldItems, wearSuccess, failureReason := user.Character.Wear(matchItem)

		if wearSuccess {

			user.Character.CancelBuffsWithFlag(buffs.Hidden)

			user.Character.RemoveItem(matchItem)

			for _, oldItem := range oldItems {
				if oldItem.ItemId != 0 {
					user.SendText(
						fmt.Sprintf(`You remove your <ansi fg="item">%s</ansi> and return it to your backpack.`, oldItem.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, user.Character.Name, oldItem.DisplayName()),
						user.UserId,
					)

					user.Character.StoreItem(oldItem)
				}
			}

			if iSpec.Subtype == items.Wearable {
				user.SendText(
					fmt.Sprintf(`You wear your <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> puts on their <ansi fg="item">%s</ansi>.`, user.Character.Name, matchItem.DisplayName()),
					user.UserId,
				)
			} else {
				user.SendText(
					fmt.Sprintf(`You wield your <ansi fg="item">%s</ansi>. You're feeling dangerous.`, matchItem.DisplayName()),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> wields their <ansi fg="item">%s</ansi>.`, user.Character.Name, matchItem.DisplayName()),
					user.UserId,
				)
			}

			// Trigger any outstanding buff onStart events
			if len(matchItem.GetSpec().WornBuffIds) > 0 {
				for _, buff := range user.Character.Buffs.List {
					if buff.OnStartWaiting {
						if _, err := scripting.TryBuffScriptEvent(`onStart`, user.UserId, 0, buff.BuffId); err == nil {
							user.Character.TrackBuffStarted(buff.BuffId)
						}
					}

				}
			}

			user.Character.Validate(true)

			events.AddToQueue(events.EquipmentChange{
				UserId:       user.UserId,
				ItemsWorn:    []items.Item{matchItem},
				ItemsRemoved: oldItems,
			})

		} else {
			if len(failureReason) == 1 {
				failureReason = fmt.Sprintf(`You can't figure out how to equip the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName())
			}
			user.SendText(
				failureReason,
			)
		}

	}

	return true, nil
}
