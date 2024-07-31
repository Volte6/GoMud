package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Remove(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if rest == "all" {
		for _, item := range user.Character.Equipment.GetAllItems() {
			r, _ := Remove(item.Name(), userId, cmdQueue)
			response.AbsorbMessages(r)
		}
		response.Handled = true
		return response, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindOnBody(rest)

	if !found || matchItem.ItemId < 1 {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't appear to be using a "%s".`, rest), true)
	} else {

		if matchItem.IsCursed() && user.Character.Health > 0 {
			if user.Character.GetSkillLevel(skills.Enchant) < 4 {
				response.SendUserMessage(userId,
					fmt.Sprintf(`You can't seem to remove your <ansi fg="item">%s</ansi>... It's <ansi fg="red-bold">CURSED!</ansi>`, matchItem.DisplayName()),
					true)

				response.Handled = true
				return response, nil
			} else {
				response.SendUserMessage(userId,
					`It's <ansi fg="red-bold">CURSED</ansi> but luckily your <ansi fg="skillname">enchant</ansi> skill level allows you to remove it.`,
					true)
			}
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		if user.Character.RemoveFromBody(matchItem) {
			response.SendUserMessage(userId,
				fmt.Sprintf(`You remove your <ansi fg="item">%s</ansi> and return it to your backpack.`, matchItem.DisplayName()),
				true)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, user.Character.Name, matchItem.DisplayName()),
				true)

			user.Character.StoreItem(matchItem)
		} else {
			response.SendUserMessage(userId,
				fmt.Sprintf(`You can't seem to remove your <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
				true)
		}

		user.Character.Validate()

	}

	response.Handled = true
	return response, nil
}
