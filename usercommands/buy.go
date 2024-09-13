package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Buy(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	if rest == "" {
		return List(rest, userId)
	}

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

	for _, mobId := range room.GetMobs(rooms.FindMerchant) {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			continue
		}

		itemNames := []string{}
		for itemId := range mob.ShopStock {
			item := items.New(itemId)
			if item.ItemId > 0 {
				itemNames = append(itemNames, item.Name())
			}
		}

		match, closeMatch := util.FindMatchIn(rest, itemNames...)
		if match == "" {
			match = closeMatch
		}

		if match == "" {
			extraSay := ""
			if len(itemNames) > 0 {
				extraSay = fmt.Sprintf(` Any interest in a <ansi fg="itemname">%s</ansi>?`, itemNames[util.Rand(len(itemNames))])
			}

			mob.Command(`say Sorry, I don't have that item right now.` + extraSay)

			response.Handled = true
			return response, nil
		}

		for itemId := range mob.ShopStock {
			item := items.New(itemId)
			if item.ItemId < 1 {
				continue
			}
			if item.Name() != match {
				continue
			}

			if user.Character.Gold < item.GetSpec().Value {

				mob.Command(`say You don't have enough gold for that.`)

				response.Handled = true
				return response, nil
			}

			user.Character.Gold -= item.GetSpec().Value
			mob.Character.Gold += item.GetSpec().Value >> 2 // They only retain 1/4th

			mob.ShopStock[itemId]--
			if mob.ShopStock[itemId] <= 0 {
				delete(mob.ShopStock, itemId)
			}

			newItm := items.New(item.ItemId)
			user.Character.StoreItem(newItm)

			iSpec := newItm.GetSpec()
			if iSpec.QuestToken != `` {

				events.AddToQueue(events.Quest{
					UserId:     user.UserId,
					QuestToken: iSpec.QuestToken,
				})

			}

			user.SendText(
				fmt.Sprintf(`You buy a <ansi fg="itemname">%s</ansi> for <ansi fg="gold">%d</ansi> gold.`, item.DisplayName(), item.GetSpec().Value),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> buys a <ansi fg="itemname">%s</ansi> from <ansi fg="mobname">%s</ansi>.`, user.Character.Name, item.DisplayName(), mob.Character.Name),
				userId,
			)

			break

		}
	}

	response.Handled = true
	return response, nil
}
