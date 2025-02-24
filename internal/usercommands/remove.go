package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/users"
)

func Remove(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	if rest == "all" {
		for _, item := range user.Character.Equipment.GetAllItems() {
			Remove(item.Name(), user, room, flags)
		}
		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindOnBody(rest)

	if !found || matchItem.ItemId < 1 {
		user.SendText(fmt.Sprintf(`You don't appear to be using a "%s".`, rest))
	} else {

		if matchItem.IsCursed() && user.Character.Health > 0 {
			if user.Character.GetSkillLevel(skills.Enchant) < 4 {
				user.SendText(
					fmt.Sprintf(`You can't seem to remove your <ansi fg="item">%s</ansi>... It's <ansi fg="red-bold">CURSED!</ansi>`, matchItem.DisplayName()),
				)

				return true, nil
			} else {
				user.SendText(
					`It's <ansi fg="red-bold">CURSED</ansi> but luckily your <ansi fg="skillname">enchant</ansi> skill level allows you to remove it.`,
				)
			}
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		if user.Character.RemoveFromBody(matchItem) {
			user.SendText(
				fmt.Sprintf(`You remove your <ansi fg="item">%s</ansi> and return it to your backpack.`, matchItem.DisplayName()),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, user.Character.Name, matchItem.DisplayName()),
				user.UserId,
			)

			user.Character.StoreItem(matchItem)
		} else {
			user.SendText(
				fmt.Sprintf(`You can't seem to remove your <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
			)
		}

		user.Character.Validate()

	}

	return true, nil
}
