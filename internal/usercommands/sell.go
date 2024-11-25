package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/users"
)

func Sell(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	item, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText("You don't have that item.")
		return true, nil
	}

	itemSpec := item.GetSpec()

	if itemSpec.ItemId < 1 {
		return true, nil
	}

	if itemSpec.QuestToken != `` {
		user.SendText("Quest items cannot be sold!")
		return true, nil
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

		user.Character.Gold += sellValue
		user.Character.RemoveItem(item)

		mob.Character.Shop.StockItem(item.ItemId)

		user.EventLog.Add(`shop`, fmt.Sprintf(`Sold your <ansi fg="itemname">%s</ansi> to <ansi fg="mobname">%s</ansi> for <ansi fg="gold">%d gold</ansi>`, item.DisplayName(), mob.Character.Name, sellValue))

		user.SendText(
			fmt.Sprintf(`You sell a <ansi fg="itemname">%s</ansi> for <ansi fg="gold">%d gold</ansi>.`, item.DisplayName(), sellValue),
		)
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> sells a <ansi fg="itemname">%s</ansi>.`, user.Character.Name, item.DisplayName()),
			user.UserId,
		)

		// Trigger lost event
		scripting.TryItemScriptEvent(`onLost`, item, user.UserId)

		break
	}

	return true, nil

}
