package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Put(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		user.SendText("Place what where?")
		return true, nil
	}

	containerName := ``
	nameSearch := ``
	for i := len(args) - 1; i >= 1; i-- {
		if len(nameSearch) > 0 {
			nameSearch = ` ` + nameSearch
		}
		nameSearch = args[i] + nameSearch

		containerName = room.FindContainerByName(nameSearch)
		if containerName != `` {
			args = args[:i]
			break
		}
	}

	if containerName == `` {
		user.SendText(`No container found by that name`)
		return true, nil
	}

	container := room.Containers[containerName]

	if container.Lock.IsLocked() {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> is locked.`, containerName))
		user.SendText(``)
		return true, nil
	}

	if len(args) < 1 {
		user.SendText("Place what where?")
		return true, nil
	}

	var item items.Item
	var itemFound bool
	goldAmt := 0

	if len(args) >= 2 && args[1] == `gold` {

		g, _ := strconv.ParseInt(args[0], 10, 32)
		goldAmt = int(g)
		if goldAmt < 0 {
			goldAmt = -1 * goldAmt
		}

	} else {

		item, itemFound = user.Character.FindInBackpack(strings.Join(args, ` `))
		if !itemFound && len(args) > 1 {
			item, itemFound = user.Character.FindInBackpack(args[0])
		}

	}

	if !itemFound && goldAmt == 0 {
		user.SendText(`You don't seem to be carrying that.`)
		return true, nil
	}

	if goldAmt > user.Character.Gold {
		user.SendText(`You don't have that much gold.`)
		return true, nil
	}

	if goldAmt > 0 {
		user.Character.Gold -= goldAmt
		container.Gold += goldAmt
		user.SendText(fmt.Sprintf(`You place <ansi fg="gold">%d gold</ansi> into the <ansi fg="container">%s</ansi>`, goldAmt, containerName))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> places some <ansi fg="gold">gold</ansi> into the <ansi fg="container">%s</ansi>`, user.Character.Name, containerName), user.UserId)
	}

	if itemFound {

		container.AddItem(item)
		user.Character.RemoveItem(item)

		user.SendText(fmt.Sprintf(`You place your <ansi fg="itemname">%s</ansi> into the <ansi fg="container">%s</ansi>`, item.DisplayName(), containerName))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> places their <ansi fg="itemname">%s</ansi> into the <ansi fg="container">%s</ansi>`, user.Character.Name, item.DisplayName(), containerName), user.UserId)

		// Enforce container size limits
		if len(container.Items) > int(configs.GetConfig().ContainerSizeMax) {

			randItemToRemove := util.Rand(len(container.Items))
			oopsItem := container.Items[randItemToRemove]

			// get all items that spawn in chests
			for _, spn := range room.SpawnInfo {
				if spn.Container == containerName && oopsItem.ItemId == spn.ItemId {
					// Don't let this one pop out
					oopsItem = item
					break
				}
			}

			container.RemoveItem(oopsItem)
			room.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> is too full and a <ansi fg="itemname">%s</ansi> falls out and onto the floor.`, containerName, oopsItem.DisplayName()))
			room.AddItem(oopsItem, false)
		}
	}

	room.Containers[containerName] = container

	return true, nil
}
