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

func Storage(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if !room.IsStorage {
		user.SendText(`You are not at a storage location.` + term.CRLFStr)
		return true, nil
	}

	itemsInStorage := user.ItemStorage.GetItems()

	if rest == `` || rest == `remove` {

		itemNames := []string{}
		for _, item := range itemsInStorage {
			itemNames = append(itemNames, item.NameComplex())
		}

		storageTxt, _ := templates.Process("character/storage", itemNames)
		user.SendText(storageTxt)

		return true, nil
	}

	if rest == `add` || rest == `remove` {
		user.SendText(fmt.Sprintf(`%s what?%s`, rest, term.CRLFStr))
		return true, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 || (args[0] != `add` && args[0] != `remove`) {
		user.SendText(`Try <ansi fg="command">help storage</ansi> for more information about storage.` + term.CRLFStr)
		return true, nil
	}

	action := args[0]
	itemName := strings.Join(args[1:], ` `)

	if action == `add` {

		spaceLeft := 20 - len(itemsInStorage)
		if spaceLeft < 1 {
			user.SendText(`You can have 20 objects in storage`)
			return true, nil
		}

		if itemName == `all` {

			for _, itm := range user.Character.GetAllBackpackItems() {
				Storage(fmt.Sprintf(`add !%d`, itm.ItemId), user)

				spaceLeft--
				if spaceLeft < 0 {
					break
				}
			}

			return true, nil
		}

		itm, found := user.Character.FindInBackpack(itemName)

		if !found {
			user.SendText(fmt.Sprintf(`You don't have a %s to add to storage.%s`, itemName, term.CRLFStr))
			return true, nil
		}

		user.Character.RemoveItem(itm)
		user.ItemStorage.AddItem(itm)

		user.SendText(fmt.Sprintf(`You placed the <ansi fg="itemname">%s</ansi> into storage.`, itm.DisplayName()))

		// Trigger lost event
		scripting.TryItemScriptEvent(`onLost`, itm, user.UserId)

	} else if action == `remove` {

		if itemName == `all` {

			for _, itm := range user.ItemStorage.GetItems() {
				Storage(fmt.Sprintf(`remove !%d`, itm.ItemId), user)
			}

			return true, nil
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
			user.SendText(fmt.Sprintf(`You don't have a %s in storage.`, itemName))
			return true, nil
		}

		if user.Character.StoreItem(itm) {

			user.ItemStorage.RemoveItem(itm)

			user.SendText(fmt.Sprintf(`You removed the <ansi fg="itemname">%s</ansi> from storage.`, itm.DisplayName()))

			scripting.TryItemScriptEvent(`onFound`, itm, user.UserId)

		} else {
			user.SendText(`You can't carry that!`)
		}

	}

	return true, nil
}
