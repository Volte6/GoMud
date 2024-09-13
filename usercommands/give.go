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

func Give(rest string, userId int) (util.MessageQueue, error) {

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

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		response.SendUserMessage(userId, "Give what? To whom?")
		response.Handled = true
		return response, nil
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
			response.SendUserMessage(userId, "You can't give a negative amount of gold.")
			response.Handled = true
			return response, nil
		}

		if giveGoldAmount > user.Character.Gold {
			response.SendUserMessage(userId, "You don't have that much gold to give.")
			response.Handled = true
			return response, nil
		}

	} else {

		var found bool = false

		// Check whether the user has an item in their inventory that matches
		giveItem, found = user.Character.FindInBackpack(giveWhat)

		if !found {
			response.SendUserMessage(userId, fmt.Sprintf("You don't have a %s to give.", giveWhat))
			response.Handled = true
			return response, nil
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

			response.SendUserMessage(userId,
				fmt.Sprintf(`You give the <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, giveItem.DisplayName(), targetUser.Character.Name),
			)
			response.SendUserMessage(targetUser.UserId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you their <ansi fg="item">%s</ansi>.`, user.Character.Name, giveItem.DisplayName()),
			)
			response.SendRoomMessage(user.Character.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> gives <ansi fg="username">%s</ansi> a <ansi fg="itemname">%s</ansi>.`, user.Character.Name, targetUser.Character.Name, giveItem.NameSimple()),
				user.UserId,
				targetUser.UserId)

			// Trigger onLost event
			if scriptResponse, err := scripting.TryItemScriptEvent(`onLost`, giveItem, userId); err == nil {
				response.AbsorbMessages(scriptResponse)
			}

			if scriptResponse, err := scripting.TryItemScriptEvent(`onFound`, giveItem, targetUser.UserId); err == nil {
				response.AbsorbMessages(scriptResponse)
			}

		} else if giveGoldAmount > 0 {

			if targetUser.UserId == user.UserId {

				response.SendUserMessage(userId,
					fmt.Sprintf(`You count out <ansi fg="gold">%d gold</ansi> and put it back in your pocket.`, giveGoldAmount),
				)
				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> counts out some <ansi fg="gold">gold</ansi> and put it back in their pocket.`, user.Character.Name),
					user.UserId)

			} else {
				targetUser.Character.Gold += giveGoldAmount
				user.Character.Gold -= giveGoldAmount

				response.SendUserMessage(userId,
					fmt.Sprintf(`You give <ansi fg="gold">%d gold</ansi> to <ansi fg="username">%s</ansi>.`, giveGoldAmount, targetUser.Character.Name),
				)
				response.SendUserMessage(targetUser.UserId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gives you <ansi fg="gold">%d gold</ansi>.`, user.Character.Name, giveGoldAmount),
				)
				response.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> gives <ansi fg="username">%s</ansi> some <ansi fg="gold">gold</ansi>.`, user.Character.Name, targetUser.Character.Name),
					user.UserId,
					targetUser.UserId)
			}
		} else {
			response.SendUserMessage(userId, "Something went wrong.")
		}

		response.Handled = true
		return response, nil

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

					response.SendUserMessage(userId,
						fmt.Sprintf(`You give <ansi fg="gold">%d gold</ansi> to <ansi fg="username">%s</ansi>.`, giveGoldAmount, m.Character.Name),
					)
					response.SendRoomMessage(room.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> gave some gold to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, m.Character.Name),
					)
				} else {

					m.Character.StoreItem(giveItem)
					user.Character.RemoveItem(giveItem)

					response.SendUserMessage(userId,
						fmt.Sprintf(`You give the <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, giveItem.DisplayName(), m.Character.Name),
					)
					response.SendRoomMessage(room.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> gave their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, user.Character.Name, giveItem.DisplayName(), m.Character.Name),
					)

					// Trigger onLost event
					if scriptResponse, err := scripting.TryItemScriptEvent(`onLost`, giveItem, userId); err == nil {
						response.AbsorbMessages(scriptResponse)
					}

				}

				if res, err := scripting.TryMobScriptEvent(`onGive`, m.InstanceId, userId, `user`, map[string]any{`gold`: giveGoldAmount, `item`: giveItem}); err == nil {
					response.AbsorbMessages(res)
					if res.Handled {
						response.Handled = true
						return response, nil
					}
				}

				m.Command(fmt.Sprintf(`emote considers the <ansi fg="itemname">%s</ansi> for a moment.`, giveItem.DisplayName()))

				m.Command(fmt.Sprintf(`gearup !%d`, giveItem.ItemId))

			} else {
				response.SendUserMessage(userId, "Something went wrong.")
			}

		}

		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId, "Who???")

	response.Handled = true
	return response, nil
}
