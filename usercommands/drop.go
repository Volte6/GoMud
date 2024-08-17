package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Drop(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		response.SendUserMessage(userId, `Drop what?`, true)

		response.Handled = true
		return response, nil
	}

	if args[0] == "all" {

		iCopies := []items.Item{}

		if user.Character.Gold > 0 {
			r, _ := Drop(fmt.Sprintf("%d gold", user.Character.Gold), userId, cmdQueue)
			response.AbsorbMessages(r)
		}

		iCopies = append(iCopies, user.Character.Items...)

		for _, item := range iCopies {
			r, _ := Drop(item.Name(), userId, cmdQueue)
			response.AbsorbMessages(r)
		}

		response.Handled = true
		return response, nil
	}

	// Drop 10 gold
	if len(args) >= 2 && args[1] == "gold" {
		g, _ := strconv.ParseInt(args[0], 10, 32)
		dropAmt := int(g)
		if dropAmt < 1 {
			response.SendUserMessage(userId, "Oops!", true)
			response.Handled = true
			return response, nil
		}

		if dropAmt > user.Character.Gold {
			response.SendUserMessage(userId, fmt.Sprintf("You don't have a %d gold to drop.", dropAmt), true)
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		room.Gold += dropAmt
		user.Character.Gold -= dropAmt

		response.SendUserMessage(userId,
			fmt.Sprintf(`You drop <ansi fg="gold">%d gold</ansi> on the floor.`, dropAmt),
			true)
		response.SendRoomMessage(room.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> drops <ansi fg="gold">%d gold</ansi>.`, user.Character.Name, dropAmt),
			true)

		response.Handled = true
		return response, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		response.SendUserMessage(userId, fmt.Sprintf("You don't have a %s to drop.", rest), true)
	} else {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		iSpec := matchItem.GetSpec()

		// Swap the item location
		user.Character.RemoveItem(matchItem)

		room.AddItem(matchItem, false)

		response.SendUserMessage(userId,
			fmt.Sprintf(`You drop the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
			true)
		response.SendRoomMessage(user.Character.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> drops their <ansi fg="item">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()),
			true)

		// If grenades are dropped, they explode and affect everyone in the room!
		if iSpec.Type == items.Grenade {
			cmdQueue.QueueRoomAction(user.Character.RoomId, user.UserId, 0, fmt.Sprintf("detonate !%d", matchItem.ItemId))
		}

		// Trigger onLost event
		if scriptResponse, err := scripting.TryItemScriptEvent(`onLost`, matchItem, userId, cmdQueue); err == nil {
			response.AbsorbMessages(scriptResponse)
		}
	}

	response.Handled = true
	return response, nil
}
