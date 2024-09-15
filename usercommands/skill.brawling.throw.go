package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
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
func Throw(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf(`user %d not found`, userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Brawling)
	handled := false

	// If they don't have a skill, act like it's not a valid command
	if skillLevel < 2 {
		return false, nil
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		user.SendText("Throw what? Where??")
		return false, nil
	}

	throwWhat := args[0]
	args = args[1:]

	throwWhere := strings.Join(args, ` `)

	itemMatch, ok := user.Character.FindInBackpack(throwWhat)
	if !ok {
		user.SendText(fmt.Sprintf(`You don't have a "%s" to throw.`, throwWhat))
		return false, nil
	}

	if !user.Character.TryCooldown(skills.Brawling.String(`throw`), 4) {
		user.SendText("You are too tired to throw objects again so soon!")
		return true, nil
	}

	targetPlayerId, targetMobId := room.FindByName(throwWhere)

	if targetMobId > 0 {
		targetMob := mobs.GetInstance(targetMobId)

		if user.Character.RemoveItem(itemMatch) {

			// Trigger onLost event
			scripting.TryItemScriptEvent(`onLost`, itemMatch, userId)

			// Tell the player they are throwing the item
			user.SendText(
				fmt.Sprintf(`You hurl the <ansi fg="itemname">%s</ansi> at <ansi fg="mobname">%s</ansi>.`, itemMatch.DisplayName(), targetMob.Character.Name),
			)

			// Tell the old room they are leaving
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="itemname">%s</ansi> at <ansi fg="mobname">%s</ansi>.`, user.Character.Name, itemMatch.DisplayName(), targetMob.Character.Name),
				userId,
			)

			// If grenades are dropped, they explode and affect everyone in the room!
			iSpec := itemMatch.GetSpec()
			if iSpec.Type == items.Grenade {

				itemMatch.SetAdjective(`exploding`, true)

				events.AddToQueue(events.RoomAction{
					RoomId:       user.Character.RoomId,
					SourceUserId: user.UserId,
					SourceMobId:  0,
					Action:       fmt.Sprintf("detonate #%d %s", targetMob.InstanceId, itemMatch.ShorthandId()),
					WaitTurns:    configs.GetConfig().TurnsPerRound() * 3,
				})

			}

			room.AddItem(itemMatch, false)

		} else {
			user.SendText(`You can't do that right now.`)
		}
		handled = true

	} else if targetPlayerId > 0 {

		targetUser := users.GetByUserId(targetPlayerId)

		user.Character.RemoveItem(itemMatch)

		// Tell the player they are throwing the item
		user.SendText(
			fmt.Sprintf(`You hurl the <ansi fg="itemname">%s</ansi> at <ansi fg="username">%s</ansi>.`, itemMatch.DisplayName(), targetUser.Character.Name),
		)

		targetUser.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> hurls their <ansi fg="itemname">%s</ansi> at you.`, itemMatch.DisplayName(), user.Character.Name),
		)

		// Tell the old room they are leaving
		room.SendText(
			fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="itemname">%s</ansi> at <ansi fg="username">%s</ansi>.`, user.Character.Name, itemMatch.DisplayName(), targetUser.Character.Name),
			userId,
			targetUser.UserId)

		// If grenades are dropped, they explode and affect everyone in the room!
		iSpec := itemMatch.GetSpec()
		if iSpec.Type == items.Grenade {

			itemMatch.SetAdjective(`exploding`, true)

			events.AddToQueue(events.RoomAction{
				RoomId:       user.Character.RoomId,
				SourceUserId: user.UserId,
				SourceMobId:  0,
				Action:       fmt.Sprintf("detonate @%d %s", targetUser.UserId, itemMatch.ShorthandId()),
				WaitTurns:    configs.GetConfig().TurnsPerRound() * 3,
			})

		}

		room.AddItem(itemMatch, false)

		handled = true

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
				user.SendText(fmt.Sprintf(`The %s exit is locked.`, exitName))
				return true, nil
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

			// Tell the player they are throwing the item
			user.SendText(
				fmt.Sprintf(`You hurl the <ansi fg="item">%s</ansi> towards the %s exit.`, itemMatch.DisplayName(), exitName),
			)

			// Tell the old room they are leaving
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, user.Character.Name, itemMatch.DisplayName(), exitName),
				userId,
			)

			// Tell the new room the item arrived
			throwToRoom.SendText(
				fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
				userId,
			)

			// If grenades are dropped, they explode and affect everyone in the room!
			iSpec := itemMatch.GetSpec()
			if iSpec.Type == items.Grenade {

				itemMatch.SetAdjective(`exploding`, true)

				events.AddToQueue(events.RoomAction{
					RoomId:       throwToRoom.RoomId,
					SourceUserId: user.UserId,
					SourceMobId:  0,
					Action:       fmt.Sprintf("detonate %s", itemMatch.ShorthandId()),
					WaitTurns:    configs.GetConfig().TurnsPerRound() * 3,
				})

			}

			throwToRoom.AddItem(itemMatch, false)

			handled = true
		}

		// Still looking for an exit... try the temp ones
		if !handled {
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

					// Tell the player they are throwing the item
					user.SendText(
						fmt.Sprintf(`You hurl the <ansi fg="item">%s</ansi> towards the %s exit.`, itemMatch.DisplayName(), tempExit.Title),
					)

					// Tell the old room they are leaving
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> throws their <ansi fg="item">%s</ansi> through the %s exit.`, user.Character.Name, itemMatch.DisplayName(), tempExit.Title),
						userId,
					)

					// Tell the new room the item arrived
					throwToRoom.SendText(
						fmt.Sprintf(`A <ansi fg="item">%s</ansi> flies through the air from %s and lands on the floor.`, itemMatch.DisplayName(), returnExitName),
						userId,
					)

					// If grenades are dropped, they explode and affect everyone in the room!
					iSpec := itemMatch.GetSpec()
					if iSpec.Type == items.Grenade {

						itemMatch.SetAdjective(`exploding`, true)

						events.AddToQueue(events.RoomAction{
							RoomId:       throwToRoom.RoomId,
							SourceUserId: user.UserId,
							SourceMobId:  0,
							Action:       fmt.Sprintf("detonate %s", itemMatch.ShorthandId()),
							WaitTurns:    configs.GetConfig().TurnsPerRound() * 3,
						})

					}

					throwToRoom.AddItem(itemMatch, false)

					handled = true

				}
			}
		}
	}

	if !handled {
		user.SendText(fmt.Sprintf(`You don't see a "%s" to throw it to.`, throwWhere))
	}

	return true, nil
}
