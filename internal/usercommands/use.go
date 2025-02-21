package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/users"
)

func Use(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	containerName := room.FindContainerByName(rest)
	if containerName != `` {

		container := room.Containers[containerName]

		if len(container.Recipes) > 0 {

			if container.Lock.IsLocked() {
				user.SendText(``)
				user.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> is locked.`, containerName))
				user.SendText(``)
				return true, nil
			}

			recipeReadyItemId := container.RecipeReady()

			if recipeReadyItemId == 0 {
				user.SendText("")
				user.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> seems to be missing something.`, containerName))
				user.SendText("")
				return true, nil
			}

			for _, removeItem := range container.Recipes[recipeReadyItemId] {
				if matchItem, found := container.FindItemById(removeItem); found {
					container.RemoveItem(matchItem)
				}
			}

			newItem := items.New(recipeReadyItemId)

			container.AddItem(newItem)
			room.Containers[containerName] = container

			room.PlaySound(`change`, `other`)

			user.SendText(``)
			user.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> produces a <ansi fg="itemname">%s</ansi>!`, containerName, newItem.DisplayName()))
			user.SendText(``)

			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> does something with the <ansi fg="container">%s</ansi>.`, user.Character.Name, containerName), user.UserId)

			return true, nil

		}

	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(rest)

	if !found {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to use.`, rest))
	} else {

		itemSpec := matchItem.GetSpec()

		if itemSpec.Subtype != items.Usable {
			user.SendText(
				fmt.Sprintf(`You can't use <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()))
			return true, nil
		}

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		user.SendText(fmt.Sprintf(`You use the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> uses their <ansi fg="itemname">%s</ansi>.`, user.Character.Name, matchItem.DisplayName()), user.UserId)

		// If no more uses, will be lost, so trigger event
		if usesLeft := user.Character.UseItem(matchItem); usesLeft < 1 {
			scripting.TryItemScriptEvent(`onLost`, matchItem, user.UserId)
		}

		for _, buffId := range itemSpec.BuffIds {

			events.AddToQueue(events.Buff{
				UserId:        user.UserId,
				MobInstanceId: 0,
				BuffId:        buffId,
			})

		}
	}

	return true, nil
}
