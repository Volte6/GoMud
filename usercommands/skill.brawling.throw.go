package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Brawling Skill
Level 2 - You can throw objects at NPCs or other rooms.
*/
func Throw(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 2 {
		return response, nil
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		response.SendUserMessage(userId, "Throw what? Where??", true)
		return response, nil
	}

	throwWhat := args[0]
	args = args[1:]

	throwWhere := strings.Join(args, ` `)

	itemMatch, ok := user.Character.FindInBackpack(throwWhat)
	if !ok {
		response.SendUserMessage(userId, fmt.Sprintf(`You don't have a "%s" to throw.`, throwWhat), true)
		return response, nil
	}

	if !user.Character.TryCooldown(skills.Brawling.String(`throw`), 4) {
		response.SendUserMessage(userId, "You are too tired to throw objects again so soon!", true)
		response.Handled = true
		return response, nil
	}

	targetPlayerId, targetMobId := room.FindByName(throwWhere)

	if targetMobId > 0 {
		targetMob := mobs.GetInstance(targetMobId)

		if user.Character.RemoveItem(itemMatch) {

			// Trigger onLost event
			if scriptResponse, err := scripting.TryItemScriptEvent(`onLost`, itemMatch, userId); err == nil {
				response.AbsorbMessages(scriptResponse)
			}

			room.AddItem(itemMatch, false)

			// Tell the player they are throwing the item
			response.SendUserMessage(userId,
				fmt.Sprintf(`You hurl the <ansi fg="itemname">%s</ansi> at <ansi fg="mobname">%s</ansi>.`, itemMatch.DisplayName(), targetMob.Character.Name),
				true)

			// Tell the old room they are leaving
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="itemname">%s</ansi> at <ansi fg="mobname">%s</ansi>.`, user.Character.Name, itemMatch.DisplayName(), targetMob.Character.Name),
				true)

			// If grenades are dropped, they explode and affect everyone in the room!
			iSpec := itemMatch.GetSpec()
			if iSpec.Type == items.Grenade {

				events.AddToQueue(events.RoomAction{
					RoomId:       user.Character.RoomId,
					SourceUserId: user.UserId,
					SourceMobId:  0,
					Action:       fmt.Sprintf("detonate #%d !%d", targetMob.InstanceId, itemMatch.ItemId),
				})

			}
		} else {
			response.SendUserMessage(userId, `You can't do that right now.`, true)
		}
		response.Handled = true

	} else if targetPlayerId > 0 {

		targetUser := users.GetByUserId(targetPlayerId)

		user.Character.RemoveItem(itemMatch)

		room.AddItem(itemMatch, false)

		// Tell the player they are throwing the item
		response.SendUserMessage(userId,
			fmt.Sprintf(`You hurl the <ansi fg="itemname">%s</ansi> at <ansi fg="username">%s</ansi>.`, itemMatch.DisplayName(), targetUser.Character.Name),
			true)

		response.SendUserMessage(targetUser.UserId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> hurls their <ansi fg="itemname">%s</ansi> at you.`, itemMatch.DisplayName(), user.Character.Name),
			true)

		// Tell the old room they are leaving
		response.SendRoomMessage(room.RoomId,
			fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="itemname">%s</ansi> at <ansi fg="username">%s</ansi>.`, user.Character.Name, itemMatch.DisplayName(), targetUser.Character.Name),
			true,
			targetUser.UserId)

		// If grenades are dropped, they explode and affect everyone in the room!
		iSpec := itemMatch.GetSpec()
		if iSpec.Type == items.Grenade {

			events.AddToQueue(events.RoomAction{
				RoomId:       user.Character.RoomId,
				SourceUserId: user.UserId,
				SourceMobId:  0,
				Action:       fmt.Sprintf("detonate @%d !%d", targetUser.UserId, itemMatch.ItemId),
			})

		}

		response.Handled = true

	} else {

		// check Exits and SecretExits for a string match. If found, move the player to that room.
		exitName, throwRoomId := room.FindExitByName(throwWhere)

		// If nothing found, consider directional aliases
		if throwRoomId == 0 {
			if alias := keywords.TryDirectionAlias(throwWhere); alias != throwWhere {
				exitName, throwRoomId = room.FindExitByName(alias)
				if throwRoomId != 0 {
					throwWhere = alias
				}
			}
		}

		if throwRoomId > 0 {

			exitInfo := room.Exits[exitName]
			if exitInfo.Lock.IsLocked() {
				response.SendUserMessage(userId, fmt.Sprintf(`The %s exit is locked.`, exitName), true)
				response.Handled = true
				return response, nil
			}

			user.Character.CancelBuffsWithFlag(buffs.Hidden)

			throwToRoom := rooms.LoadRoom(throwRoomId)
			returnExitName := throwToRoom.FindExitTo(user.Character.RoomId)

			if len(returnExitName) < 1 {
				returnExitName = "somewhere"
			} else {
				returnExitName = fmt.Sprintf("the %s exit", returnExitName)
			}

			user.Character.RemoveItem(itemMatch)
			throwToRoom.AddItem(itemMatch, false)

			// Tell the player they are throwing the item
			response.SendUserMessage(userId,
				fmt.Sprintf(`You hurl the <ansi fg="item">%s</ansi> towards the %s exit.`, itemMatch.DisplayName(), exitName),
				true)

			// Tell the old room they are leaving
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, user.Character.Name, itemMatch.DisplayName(), exitName),
				true)

			// Tell the new room the item arrived
			response.SendRoomMessage(throwToRoom.RoomId,
				fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
				true)

			// If grenades are dropped, they explode and affect everyone in the room!
			iSpec := itemMatch.GetSpec()
			if iSpec.Type == items.Grenade {

				events.AddToQueue(events.RoomAction{
					RoomId:       throwToRoom.RoomId,
					SourceUserId: user.UserId,
					SourceMobId:  0,
					Action:       fmt.Sprintf("detonate !%d", itemMatch.ItemId),
				})

			}

			response.Handled = true
		}

		// Still looking for an exit... try the temp ones
		if !response.Handled {
			if len(room.ExitsTemp) > 0 {
				// See if there's a close match
				exitNames := make([]string, 0, len(room.ExitsTemp))
				for exitName := range room.ExitsTemp {
					exitNames = append(exitNames, exitName)
				}

				exactMatch, closeMatch := util.FindMatchIn(throwWhere, exitNames...)

				var tempExit rooms.TemporaryRoomExit
				var tempExitFound bool = false
				if len(exactMatch) > 0 {
					tempExit = room.ExitsTemp[exactMatch]
					tempExitFound = true
				} else if len(closeMatch) > 0 && len(rest) >= 3 {
					tempExit = room.ExitsTemp[closeMatch]
					tempExitFound = true
				}

				if tempExitFound {

					user.Character.CancelBuffsWithFlag(buffs.Hidden)

					// do something with tempExit
					throwToRoom := rooms.LoadRoom(tempExit.RoomId)
					returnExitName := throwToRoom.FindExitTo(user.Character.RoomId)

					if len(returnExitName) < 1 {
						returnExitName = "somewhere"
					} else {
						returnExitName = fmt.Sprintf("the %s exit", returnExitName)
					}

					user.Character.RemoveItem(itemMatch)
					throwToRoom.AddItem(itemMatch, false)

					// Tell the player they are throwing the item
					response.SendUserMessage(userId,
						fmt.Sprintf(`You hurl the <ansi fg="item">%s</ansi> towards the %s exit.`, itemMatch.DisplayName(), tempExit.Title),
						true)

					// Tell the old room they are leaving
					response.SendRoomMessage(room.RoomId,
						fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, user.Character.Name, itemMatch.DisplayName(), tempExit.Title),
						true)

					// Tell the new room the item arrived
					response.SendRoomMessage(tempExit.RoomId,
						fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
						true)

					// If grenades are dropped, they explode and affect everyone in the room!
					iSpec := itemMatch.GetSpec()
					if iSpec.Type == items.Grenade {

						events.AddToQueue(events.RoomAction{
							RoomId:       throwToRoom.RoomId,
							SourceUserId: user.UserId,
							SourceMobId:  0,
							Action:       fmt.Sprintf("detonate !%d", itemMatch.ItemId),
						})

					}

					response.Handled = true

				}
			}
		}
	}

	if !response.Handled {
		response.Handled = true
		response.SendUserMessage(userId, fmt.Sprintf(`You don't see a "%s" to throw it to.`, throwWhere), true)
	}

	return response, nil
}
