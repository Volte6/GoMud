package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Give(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		user.SendText("Give what? To whom?")
		return true, nil
	}

	var giveWho string = args[len(args)-1]
	args = args[:len(args)-1]
	var giveWhat string = strings.Join(args, " ")

	var giveItem items.Item = items.Item{}
	var giveGoldAmount int = 0

	if len(giveWhat) > 4 && giveWhat[len(giveWhat)-4:] == "gold" {

		g, _ := strconv.ParseInt(giveWhat[0:len(giveWhat)-5], 10, 32)
		giveGoldAmount = int(g)

		if giveGoldAmount < 0 {
			user.SendText("You can't give a negative amount of gold.")
			return true, nil
		}

		if giveGoldAmount > user.Character.Gold {
			user.SendText("You don't have that much gold to give.")
			return true, nil
		}

	} else {

		var found bool = false

		// Check whether the user has an item in their inventory that matches
		giveItem, found = user.Character.FindInBackpack(giveWhat)

		if !found {
			user.SendText(fmt.Sprintf("You don't have a %s to give.", giveWhat))
			return true, nil
		}

	}

	playerId, mobId := room.FindByName(giveWho)

	if playerId > 0 {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		targetUser := users.GetByUserId(playerId)

		// Swap the item location
		if giveItem.ItemId > 0 {
			targetUser.Character.StoreItem(giveItem)
			user.Character.RemoveItem(giveItem)

			iSpec := giveItem.GetSpec()
			if iSpec.QuestToken != `` {

				events.AddToQueue(events.Quest{
					UserId:     targetUser.UserId,
					QuestToken: iSpec.QuestToken,
				})

			}

			user.SendText(
				fmt.Sprintf(`You give the <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, giveItem.DisplayName(), targetUser.Character.Name),
			)
			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you their <ansi fg="item">%s</ansi>.`, user.Character.Name, giveItem.DisplayName()),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives <ansi fg="username">%s</ansi> a <ansi fg="itemname">%s</ansi>.`, user.Character.Name, targetUser.Character.Name, giveItem.NameSimple()),
				user.UserId,
				targetUser.UserId)

			// Trigger onLost event
			scripting.TryItemScriptEvent(`onLost`, giveItem, user.UserId)

			scripting.TryItemScriptEvent(`onFound`, giveItem, targetUser.UserId)

		} else if giveGoldAmount > 0 {

			if targetUser.UserId == user.UserId {

				user.SendText(
					fmt.Sprintf(`You count out <ansi fg="gold">%d gold</ansi> and put it back in your pocket.`, giveGoldAmount),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> counts out some <ansi fg="gold">gold</ansi> and put it back in their pocket.`, user.Character.Name),
					user.UserId)

			} else {
				targetUser.Character.Gold += giveGoldAmount
				user.Character.Gold -= giveGoldAmount

				user.SendText(
					fmt.Sprintf(`You give <ansi fg="gold">%d gold</ansi> to <ansi fg="username">%s</ansi>.`, giveGoldAmount, targetUser.Character.Name),
				)
				targetUser.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you <ansi fg="gold">%d gold</ansi>.`, user.Character.Name, giveGoldAmount),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gives <ansi fg="username">%s</ansi> some <ansi fg="gold">gold</ansi>.`, user.Character.Name, targetUser.Character.Name),
					user.UserId,
					targetUser.UserId)
			}
		} else {
			user.SendText("Something went wrong.")
		}

		return true, nil

	}

	//
	// Look for an NPC
	//
	if mobId > 0 {

		user.Character.CancelBuffsWithFlag(buffs.Hidden)

		m := mobs.GetInstance(mobId)

		if m != nil {

			// Swap the item location
			if giveItem.ItemId > 0 || giveGoldAmount > 0 {

				if giveGoldAmount > 0 {
					m.Character.Gold += giveGoldAmount
					user.Character.Gold -= giveGoldAmount

					user.SendText(
						fmt.Sprintf(`You give <ansi fg="gold">%d gold</ansi> to <ansi fg="username">%s</ansi>.`, giveGoldAmount, m.Character.Name),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> gave some gold to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, m.Character.Name),
						user.UserId,
					)
				} else {

					m.Character.StoreItem(giveItem)
					user.Character.RemoveItem(giveItem)

					user.SendText(
						fmt.Sprintf(`You give the <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, giveItem.DisplayName(), m.Character.Name),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> gave their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, giveItem.DisplayName(), m.Character.Name),
						user.UserId,
					)

					// Trigger onLost event
					scripting.TryItemScriptEvent(`onLost`, giveItem, user.UserId)

				}

				if handled, err := scripting.TryMobScriptEvent(`onGive`, m.InstanceId, user.UserId, `user`, map[string]any{`gold`: giveGoldAmount, `item`: giveItem}); err == nil {
					if handled {
						return true, nil
					}
				}

				if giveGoldAmount > 0 {
					m.Command(`emote counts his gold coins and chuckles a bit.`)
				} else {
					m.Command(fmt.Sprintf(`emote considers the <ansi fg="itemname">%s</ansi> for a moment.`, giveItem.DisplayName()))
					m.Command(fmt.Sprintf(`gearup !%d`, giveItem.ItemId))
				}
			} else {
				user.SendText("Something went wrong.")
			}

		}

		return true, nil
	}

	//
	// Look for any pets in the room
	//
	petUserId := room.FindByPetName(giveWho)
	if petUserId > 0 {

		petUser := users.GetByUserId(petUserId)
		if petUser == nil {
			user.SendText("Who???")
			return true, nil
		}

		if giveGoldAmount > 0 {
			room.SendText(fmt.Sprintf(`What would %s do with <ansi fg="gold">%d gold</ansi>?`, petUser.Character.Pet.DisplayName(), giveGoldAmount))
			return true, nil
		}

		user.SendText(fmt.Sprintf(`You give the <ansi fg="itemname">%s</ansi> to %s.`, giveItem.DisplayName(), petUser.Character.Pet.DisplayName()))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> gives their <ansi fg="itemname">%s</ansi> to %s...`, user.Character.Name, giveItem.DisplayName(), petUser.Character.Pet.DisplayName()), user.UserId)

		user.Character.RemoveItem(giveItem)

		if len(petUser.Character.Pet.Items) >= petUser.Character.Pet.Capacity || !petUser.Character.Pet.StoreItem(giveItem) {
			room.SendText(fmt.Sprintf(`%s throws the <ansi fg="itemname">%s</ansi> onto the ground.`, petUser.Character.Pet.DisplayName(), giveItem.DisplayName()))
			room.AddItem(giveItem, false)
		}

		return true, nil
	}

	user.SendText("Who???")

	return true, nil
}
