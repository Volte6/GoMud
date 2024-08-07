package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Storage(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	if !room.IsStorage {
		response.SendUserMessage(userId, `You are not at a storage location.`+term.CRLFStr, true)
		response.Handled = true
		return response, nil
	}

	itemsInStorage := user.ItemStorage.GetItems()

	if rest == `` || rest == `remove` {

		itemNames := []string{}
		for _, item := range itemsInStorage {
			itemNames = append(itemNames, item.NameComplex())
		}

		storageTxt, _ := templates.Process("character/storage", itemNames)
		response.SendUserMessage(userId, storageTxt, false)

		response.Handled = true
		return response, nil
	}

	if rest == `add` || rest == `remove` {
		response.SendUserMessage(userId, fmt.Sprintf(`%s what?%s`, rest, term.CRLFStr), true)
		response.Handled = true
		return response, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 || (args[0] != `add` && args[0] != `remove`) {
		response.SendUserMessage(userId, `Try <ansi fg="command">help storage</ansi> for more information about storage.`+term.CRLFStr, true)
		response.Handled = true
		return response, nil
	}

	action := args[0]
	itemName := strings.Join(args[1:], ` `)

	if action == `add` {

		spaceLeft := 20 - len(itemsInStorage)
		if spaceLeft < 1 {
			response.SendUserMessage(userId, `You can have 20 objects in storage`, true)
			response.Handled = true
			return response, nil
		}

		if itemName == `all` {

			for _, itm := range user.Character.GetAllBackpackItems() {
				r, _ := Storage(fmt.Sprintf(`add !%d`, itm.ItemId), userId, cmdQueue)
				response.AbsorbMessages(r)
				spaceLeft--
				if spaceLeft < 0 {
					break
				}
			}

			response.Handled = true
			return response, nil
		}

		itm, found := user.Character.FindInBackpack(itemName)

		if !found {
			response.SendUserMessage(userId, fmt.Sprintf(`You don't have a %s to add to storage.%s`, itemName, term.CRLFStr), true)
			response.Handled = true
			return response, nil
		}

		user.Character.RemoveItem(itm)
		user.ItemStorage.AddItem(itm)

		response.SendUserMessage(userId, fmt.Sprintf(`You placed the <ansi fg="itemname">%s</ansi> into storage.`, itm.DisplayName()), true)

		// Trigger lost event
		if scriptResponse, err := scripting.TryItemScriptEvent(`onLost`, itm, userId, cmdQueue); err == nil {
			response.AbsorbMessages(scriptResponse)
		}

	} else if action == `remove` {

		if itemName == `all` {

			for _, itm := range user.ItemStorage.GetItems() {
				r, _ := Storage(fmt.Sprintf(`remove !%d`, itm.ItemId), userId, cmdQueue)
				response.AbsorbMessages(r)
			}

			response.Handled = true
			return response, nil
		}

		var itm items.Item
		var found bool = false
		itmIdx, _ := strconv.Atoi(itemName)

		if itmIdx > 0 {
			itmIdx -= 1
			for i, storageItm := range itemsInStorage {
				if itmIdx == i {
					itm = storageItm
					found = true
					break
				}
			}

		} else {
			itm, found = user.ItemStorage.FindItem(itemName)
		}

		if !found {
			response.SendUserMessage(userId, fmt.Sprintf(`You don't have a %s in storage.`, itemName), true)
			response.Handled = true
			return response, nil
		}

		if user.Character.StoreItem(itm) {

			user.ItemStorage.RemoveItem(itm)

			response.SendUserMessage(userId, fmt.Sprintf(`You removed the <ansi fg="itemname">%s</ansi> from storage.`, itm.DisplayName()), true)

			if scriptResponse, err := scripting.TryItemScriptEvent(`onFound`, itm, userId, cmdQueue); err == nil {
				response.AbsorbMessages(scriptResponse)
			}

		} else {
			response.SendUserMessage(userId, `You can't carry that!`, true)
		}

	}

	response.Handled = true
	return response, nil
}
