package mobcommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Remove(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.HasBuffFlag(buffs.PermaGear) {
		mob.Command(`emote struggles with their gear for a while, then gives up.`)
		return true, nil
	}

	if rest == "all" {
		removedItems := []items.Item{}
		for _, item := range mob.Character.Equipment.GetAllItems() {
			Remove(item.Name(), mob, room)
			removedItems = append(removedItems, item)
		}

		events.AddToQueue(events.EquipmentChange{
			MobInstanceId: mob.InstanceId,
			ItemsRemoved:  removedItems,
		})

		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindOnBody(rest)

	if found && matchItem.ItemId > 0 {

		if mob.Character.RemoveFromBody(matchItem) {

			mob.Character.CancelBuffsWithFlag(buffs.Hidden)

			room.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, mob.Character.Name, matchItem.DisplayName()),
			)

			mob.Character.StoreItem(matchItem)
		}

		mob.Character.Validate()

		events.AddToQueue(events.EquipmentChange{
			MobInstanceId: mob.InstanceId,
			ItemsRemoved:  []items.Item{matchItem},
		})

	}

	return true, nil
}
