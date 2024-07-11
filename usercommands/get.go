package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Get(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		response.SendUserMessage(userId, "Get what?", true)
		response.Handled = true
		return response, nil
	}

	if args[0] == "all" {
		if room.Gold > 0 {
			r, _ := Get(`gold`, userId, cmdQueue)
			response.AbsorbMessages(r)
		}

		if len(room.Items) > 0 {
			iCopies := append([]items.Item{}, room.Items...)

			for _, item := range iCopies {
				r, _ := Get(item.Name(), userId, cmdQueue)
				response.AbsorbMessages(r)
			}
		}

		response.Handled = true
		return response, nil
	}

	getFromStash := false
	containerName := ``

	if len(args) >= 2 {
		// Detect "stash" or "from stash" at end and remove it
		if args[len(args)-1] == "stash" {
			getFromStash = true
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

		if args[len(args)-1] == "ground" {
			getFromStash = false
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

		containerName = room.FindContainerByName(args[len(args)-1])
		if containerName != `` {
			getFromStash = false
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

	}

	if containerName != `` {
		container := room.Containers[containerName]

		goldName := `gold`
		if args[0] == goldName || (len(args[0]) < 5 && goldName[0:len(args[0])-1] == args[0]) {

			if container.Gold < 1 {
				response.SendUserMessage(userId, "There's no gold to grab.", true)
			} else {

				user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

				goldAmt := container.Gold
				user.Character.Gold += goldAmt
				container.Gold -= goldAmt
				room.Containers[containerName] = container

				response.SendUserMessage(userId,
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi> from the <ansi fg="container">%s</ansi>.`, goldAmt, containerName),
					true)
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi> from the <ansi fg="container">%s</ansi>.`, user.Character.Name, containerName),
					true)
			}

			response.Handled = true
			return response, nil
		}

		matchItem, found := container.FindItem(rest)

		if !found {
			response.SendUserMessage(userId, fmt.Sprintf(`You don't see a %s in the <ansi fg="container">%s</ansi>.`, rest, containerName), true)
		} else {

			user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

			// Trigger onFound event
			if user.Character.StoreItem(matchItem) {

				// Swap the item location
				container.RemoveItem(matchItem)
				room.Containers[containerName] = container

				iSpec := matchItem.GetSpec()
				if iSpec.QuestToken != `` {
					cmdQueue.QueueQuest(user.UserId, iSpec.QuestToken)
				}

				response.SendUserMessage(userId,
					fmt.Sprintf(`You take the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>.`, matchItem.DisplayName(), containerName),
					true)
				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>...`, user.Character.Name, matchItem.DisplayName(), containerName),
					true)

				if scriptResponse, err := scripting.TryItemScriptEvent(`onFound`, matchItem, userId, cmdQueue); err == nil {
					response.AbsorbMessages(scriptResponse)
				}

			} else {
				response.SendUserMessage(userId,
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
					true)
			}

		}

	} else {

		goldName := `gold`
		if args[0] == goldName || (len(args[0]) < 5 && goldName[0:len(args[0])-1] == args[0]) {

			if room.Gold < 1 {
				response.SendUserMessage(userId, "There's no gold to grab.", true)
			} else {

				user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

				goldAmt := room.Gold
				user.Character.Gold += goldAmt
				room.Gold -= goldAmt

				response.SendUserMessage(userId,
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi>.`, goldAmt),
					true)
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi>.`, user.Character.Name),
					true)
			}

			response.Handled = true
			return response, nil
		}

		// Check whether the user has an item in their inventory that matches
		matchItem, found := room.FindOnFloor(rest, getFromStash)

		if !found {
			response.SendUserMessage(userId, fmt.Sprintf("You don't see a %s around.", rest), true)
		} else {

			user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

			if user.Character.StoreItem(matchItem) {

				// Swap the item location
				room.RemoveItem(matchItem, getFromStash)

				iSpec := matchItem.GetSpec()
				if iSpec.QuestToken != `` {
					cmdQueue.QueueQuest(user.UserId, iSpec.QuestToken)
				}

				response.SendUserMessage(userId,
					fmt.Sprintf(`You pick up the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
					true)
				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()),
					true)

				if scriptResponse, err := scripting.TryItemScriptEvent(`onFound`, matchItem, userId, cmdQueue); err == nil {
					response.AbsorbMessages(scriptResponse)
				}

			} else {
				response.SendUserMessage(userId,
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
					true)
			}
		}

	}

	response.Handled = true
	return response, nil
}
