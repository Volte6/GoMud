package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Sell(rest string, userId int) (util.MessageQueue, error) {

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

	item, found := user.Character.FindInBackpack(rest)
	if !found {
		user.SendText("You don't have that item.")
		response.Handled = true
		return response, nil
	}

	itemSpec := item.GetSpec()
	if itemSpec.ItemId < 1 {
		response.Handled = true
		return response, nil
	}

	if itemSpec.QuestToken != `` {
		user.SendText("Quest items cannot be sold!")
		response.Handled = true
		return response, nil
	}

	for _, mobId := range room.GetMobs(rooms.FindMerchant) {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		if item.IsSpecial() {

			mob.Command(`say I'm afraid I don't buy those.`)

			continue
		}

		sellValue := mob.GetSellPrice(item)

		if sellValue <= 0 {

			mob.Command(`say I'm not interested in that.`)

			continue
		}

		if sellValue > mob.Character.Gold {

			mob.Command(`say I'm low on funds right now. Maybe later.`)

			continue
		}

		mob.Character.Gold -= sellValue
		user.Character.Gold += sellValue
		user.Character.RemoveItem(item)

		if _, ok := mob.ShopStock[item.ItemId]; !ok {
			mob.ShopStock[item.ItemId] = 1
		} else {
			mob.ShopStock[item.ItemId]++
		}

		user.SendText(
			fmt.Sprintf(`You sell a <ansi fg="itemname">%s</ansi> for <ansi fg="gold">%d</ansi> gold.`, item.DisplayName(), sellValue),
		)
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> sells a <ansi fg="itemname">%s</ansi>.`, user.Character.Name, item.DisplayName()),
			userId,
		)

		// Trigger lost event
		if scriptResponse, err := scripting.TryItemScriptEvent(`onLost`, item, userId); err == nil {
			response.AbsorbMessages(scriptResponse)
		}

		break
	}

	response.Handled = true
	return response, nil
}
