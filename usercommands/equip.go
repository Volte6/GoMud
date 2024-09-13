package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Equip(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if rest == "all" {
		return Gearup(``, userId)
		itemCopies := append([]items.Item{}, user.Character.Items...)
		for _, item := range itemCopies {
			iSpec := item.GetSpec()
			if iSpec.Subtype == items.Wearable || iSpec.Type == items.Weapon {
				r, _ := Equip(item.Name(), userId)
				response.AbsorbMessages(r)
			}
		}
		response.Handled = true
		return response, nil
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
			response.Handled = true
			return response, nil
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
						userId,
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
					userId,
				)
			} else {
				user.SendText(
					fmt.Sprintf(`You wield your <ansi fg="item">%s</ansi>. You're feeling dangerous.`, matchItem.DisplayName()),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> wields their <ansi fg="item">%s</ansi>.`, user.Character.Name, matchItem.DisplayName()),
					userId,
				)
			}

			// Trigger any outstanding buff onStart events
			if len(matchItem.GetSpec().WornBuffIds) > 0 {
				for _, buff := range user.Character.Buffs.List {
					if !buff.OnStartEvent {
						if scriptResponse, err := scripting.TryBuffScriptEvent(`onStart`, user.UserId, 0, buff.BuffId); err == nil {
							response.AbsorbMessages(scriptResponse)
							user.Character.TrackBuffStarted(buff.BuffId)
						}
					}
				}
			}

			user.Character.Validate(true)
		} else {
			if len(failureReason) == 1 {
				failureReason = fmt.Sprintf(`You can't figure out how to equip the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName())
			}
			user.SendText(
				failureReason,
			)
		}

	}

	response.Handled = true
	return response, nil
}
