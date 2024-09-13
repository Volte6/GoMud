package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Drink(rest string, userId int) (util.MessageQueue, error) {

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

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't have a "%s" to drink.`, rest))
	} else {

		itemSpec := matchItem.GetSpec()

		if itemSpec.Subtype != items.Drinkable {
			response.SendUserMessage(userId,
				fmt.Sprintf(`You can't drink <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
			)
			response.Handled = true
			return response, nil
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		user.Character.UseItem(matchItem)

		response.SendUserMessage(userId, fmt.Sprintf(`You drink the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()))
		response.SendRoomMessage(room.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> drinks <ansi fg="itemname">%s</ansi>.`, user.Character.Name, matchItem.DisplayName()))

		for _, buffId := range itemSpec.BuffIds {

			events.AddToQueue(events.Buff{
				UserId:        user.UserId,
				MobInstanceId: 0,
				BuffId:        buffId,
			})

		}
	}

	response.Handled = true
	return response, nil
}
