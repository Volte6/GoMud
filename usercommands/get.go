package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Get(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		user.SendText("Get what?")
		return true, nil
	}

	if args[0] == "all" {
		if room.Gold > 0 {
			Get(`gold`, user, room)
		}

		if len(room.Items) > 0 {
			iCopies := append([]items.Item{}, room.Items...)

			for _, item := range iCopies {
				Get(item.Name(), user, room)
			}
		}

		return true, nil
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
				user.SendText("There's no gold to grab.")
			} else {

				user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

				goldAmt := container.Gold
				user.Character.Gold += goldAmt
				container.Gold -= goldAmt
				room.Containers[containerName] = container

				user.SendText(
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi> from the <ansi fg="container">%s</ansi>.`, goldAmt, containerName),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi> from the <ansi fg="container">%s</ansi>.`, user.Character.Name, containerName),
					user.UserId,
				)
			}

			return true, nil
		}

		matchItem, found := container.FindItem(rest)

		if !found {
			user.SendText(fmt.Sprintf(`You don't see a %s in the <ansi fg="container">%s</ansi>.`, rest, containerName))
		} else {

			user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

			// Trigger onFound event
			if user.Character.StoreItem(matchItem) {

				// Swap the item location
				container.RemoveItem(matchItem)
				room.Containers[containerName] = container

				iSpec := matchItem.GetSpec()
				if iSpec.QuestToken != `` {

					events.AddToQueue(events.Quest{
						UserId:     user.UserId,
						QuestToken: iSpec.QuestToken,
					})

				}

				user.SendText(
					fmt.Sprintf(`You take the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>.`, matchItem.DisplayName(), containerName),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>...`, user.Character.Name, matchItem.DisplayName(), containerName),
					user.UserId,
				)

				scripting.TryItemScriptEvent(`onFound`, matchItem, user.UserId)

			} else {
				user.SendText(
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
				)
			}

		}

	} else {

		goldName := `gold`
		if args[0] == goldName || (len(args[0]) < 5 && goldName[0:len(args[0])-1] == args[0]) {

			if room.Gold < 1 {
				user.SendText("There's no gold to grab.")
			} else {

				user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

				goldAmt := room.Gold
				user.Character.Gold += goldAmt
				room.Gold -= goldAmt

				user.SendText(
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi>.`, goldAmt),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi>.`, user.Character.Name),
					user.UserId,
				)
			}

			return true, nil
		}

		// Check whether the user has an item in their inventory that matches
		matchItem, found := room.FindOnFloor(rest, getFromStash)

		// Check if user is specifying an item they stashed
		if !found && !getFromStash {
			stashItemMatch, stashFound := room.FindOnFloor(rest, true)
			if stashFound && stashItemMatch.StashedBy == user.UserId {
				found = true
				getFromStash = true
				matchItem = stashItemMatch
			}
		}

		if !found {
			user.SendText(fmt.Sprintf("You don't see a %s around.", rest))
		} else {

			if matchItem.HasAdjective(`exploding`) {
				user.SendText(`You can't pick that up, it's about to explode!`)
				return true, nil
			}

			user.Character.CancelBuffsWithFlag(buffs.Hidden) // No longer sneaking

			// If it was in the stash, remove the stash owner tag
			if getFromStash {
				matchItem.StashedBy = 0
			}

			if user.Character.StoreItem(matchItem) {

				// Swap the item location
				room.RemoveItem(matchItem, getFromStash)

				iSpec := matchItem.GetSpec()
				if iSpec.QuestToken != `` {

					events.AddToQueue(events.Quest{
						UserId:     user.UserId,
						QuestToken: iSpec.QuestToken,
					})

				}

				if getFromStash {
					user.SendText(
						fmt.Sprintf(`You dig out the <ansi fg="itemname">%s</ansi> from where it was stashed.`, matchItem.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> digs around in the area and picks something up...`, user.Character.Name),
						user.UserId,
					)
				} else {
					user.SendText(
						fmt.Sprintf(`You pick up the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi>...`, user.Character.Name, matchItem.DisplayName()),
						user.UserId,
					)
				}
				scripting.TryItemScriptEvent(`onFound`, matchItem, user.UserId)

			} else {
				user.SendText(
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
				)
			}
		}

	}

	return true, nil
}
