package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Offer(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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
		response.SendUserMessage(user.UserId, "You don't have that item.", true)
		response.Handled = true
		return response, nil
	}

	itemSpec := item.GetSpec()
	if itemSpec.ItemId < 1 {
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
			cmdQueue.QueueCommand(0, mobId, "say I'm afraid I don't buy those.")
			continue
		}

		sellValue := mob.GetSellPrice(item)

		if sellValue <= 0 {
			cmdQueue.QueueCommand(0, mobId, "say I'm not interested in that.")
			continue
		}

		cmdQueue.QueueCommand(0, mobId, fmt.Sprintf(`say I can give you <ansi fg="gold">%d gold</ansi> for that <ansi fg="itemname">%s</ansi>.`, sellValue, item.DisplayName()))

		break
	}

	response.Handled = true
	return response, nil
}
