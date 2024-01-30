package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Go(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {
	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Aggro != nil {
		response.SendUserMessage(userId, "You can't do that! You are in combat!", true)
		response.Handled = true
		return response, nil
	}

	// If has a buff that prevents combat, skip the player
	if user.Character.HasBuffFlag(buffs.NoMovement) {
		response.SendUserMessage(userId, "You can't do that!", true)
		response.Handled = true
		return response, nil
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	exitName, goRoomId := room.FindExitByName(rest)

	exitInfo := room.Exits[exitName]
	if exitInfo.Lock.IsLocked() {

		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, exitName)

		hasKey, hasSequence := user.Character.HasKey(lockId, int(room.Exits[exitName].Lock.Difficulty))

		lockpickItm := items.Item{}
		// Only look for a lockpick kit if they know the sequence
		if hasSequence {
			for _, itm := range user.Character.GetAllBackpackItems() {
				if itm.GetSpec().Type == items.Lockpicks {
					lockpickItm = itm
					break
				}
			}
		}

		if lockpickItm.ItemId > 0 && hasSequence {

			response.SendUserMessage(userId, `You know this lock well, you quickly pick it.`, true)
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> quickly picks the lock on the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
				true,
				userId)

			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo

		} else if hasKey {
			response.SendUserMessage(userId, fmt.Sprintf(`You use the key on your key ring to unlock the <ansi fg="exit">%s</ansi> exit.`, exitName), true)
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
				true,
				userId)

			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo
		} else {

			// check for a key item on their person
			if backpackKeyItm, hasBackpackKey := user.Character.FindKeyInBackpack(lockId); hasBackpackKey {

				itmSpec := backpackKeyItm.GetSpec()

				response.SendUserMessage(userId, fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to unlock the <ansi fg="exit">%s</ansi> exit, and add it to your key ring for the future.`, itmSpec.Name, exitName), true)
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
					true,
					userId)

				// Key entries look like:
				// "key-<roomid>-<exitname>": "<itemid>"
				user.Character.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, backpackKeyItm.ItemId))
				user.Character.RemoveItem(backpackKeyItm)

				exitInfo.Lock.SetUnlocked()
				room.Exits[exitName] = exitInfo

			}

			if exitInfo.Lock.IsLocked() {
				response.SendUserMessage(userId, `There's a lock preventing you from going that way. You'll need a <ansi fg="item">Key</ansi> or to <ansi fg="command">pick</ansi> the lock with <ansi fg="item">lockpicks</ansi>.`, true)
				response.Handled = true
				return response, nil
			}
		}

	}

	if goRoomId > 0 || exitName != `` {
		// It does so we won't need to continue down the logic after this chunk
		response.Handled = true

		// Load current room details
		destRoom := rooms.LoadRoom(goRoomId)
		if destRoom == nil {
			return response, fmt.Errorf(`room %d not found`, goRoomId)
		}

		// Grab the exit in the target room that leads to this room (if any)
		enterFromExit := destRoom.FindExitTo(room.RoomId)

		if len(enterFromExit) < 1 {
			enterFromExit = "somewhere"
		} else {

			// Entering through the other side unlocks this side
			exitInfo := destRoom.Exits[enterFromExit]
			if exitInfo.Lock.IsLocked() {
				exitInfo.Lock.SetUnlocked()
				destRoom.Exits[enterFromExit] = exitInfo
			}

			enterFromExit = fmt.Sprintf(`the <ansi fg="exit">%s</ansi>`, enterFromExit)
		}

		if err := rooms.MoveToRoom(user.UserId, destRoom.RoomId); err != nil {
			response.SendUserMessage(userId, "Oops, couldn't move there!", true)
		} else {
			// Tell the player they are moving
			if isSneaking {
				response.SendUserMessage(userId, fmt.Sprintf(`You <ansi fg="black" bold="true">sneak</ansi> towards the %s exit.`, exitName), true)
			} else {
				response.SendUserMessage(userId, fmt.Sprintf(`You head towards the <ansi fg="exit">%s</ansi> exit.`, exitName), true)

				// Tell the old room they are leaving
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> leaves towards the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
					true)
				// Tell the new room they have arrived
				response.SendRoomMessage(destRoom.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> enters from %s.`, user.Character.Name, enterFromExit),
					true)
			}

			if currentParty := parties.Get(userId); currentParty != nil {

				if currentParty.IsLeader(userId) {

					for _, partyMemberId := range currentParty.UserIds {
						if partyMemberId == userId {
							continue
						}
						if partyUser := users.GetByUserId(partyMemberId); partyUser != nil {
							if partyUser.Character.RoomId == room.RoomId {
								response.SendUserMessage(partyMemberId, `You follow the party leader.`, true)
								cmdQueue.QueueCommand(partyMemberId, 0, rest)
							}
						}
					}

				}
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				mob := mobs.GetInstance(instId)
				if mob == nil {
					continue
				}
				if mob.Character.IsCharmed(userId) { // Charmed mobs follow
					cmdQueue.QueueCommand(0, instId, rest)
				}
			}

			if !isSneaking {
				//
				// When leaving a room, mobs who were attacking may follow
				//
				mobInstanceIds := room.GetMobs(rooms.FindFightingPlayer)
				for _, mobInstanceId := range mobInstanceIds {
					mob := mobs.GetInstance(mobInstanceId)
					if mob == nil {
						continue
					}

					if mob.Character.Aggro == nil || mob.Character.Aggro.UserId != user.UserId {
						continue
					}

					speedDelta := mob.Character.Stats.Speed.Value - user.Character.Stats.Speed.Value
					if speedDelta < 1 {
						speedDelta = 1
					}

					// Chance that a mob follows the player
					targetVal := 20 + mob.Character.Stats.Perception.Value + speedDelta

					roll := util.Rand(100)

					util.LogRoll(`Mob Follow`, roll, targetVal)

					if roll >= targetVal {
						continue
					}

					cmdQueue.QueueCommand(0, mob.InstanceId, rest)
				}

				//
				// When entering a room, mobs might be waiting to attack
				//
				mobInstanceIds = destRoom.GetMobs(rooms.FindAll)
				for _, mobInstanceId := range mobInstanceIds {
					mob := mobs.GetInstance(mobInstanceId)
					if mob == nil {
						continue
					}
					if mob.Character.Aggro != nil {
						continue
					}
					if mob.Character.IsCharmed() {
						continue
					}

					isHostile := mob.Hostile // Is it automatically hostile?
					if !isHostile {
						for _, groupName := range mob.Groups {
							if mobs.IsHostile(groupName, user.UserId) {
								isHostile = true
								break
							}
						}
						if !isHostile { // is it still not hostile?
							continue
						}
					}

					response.SendUserMessage(user.UserId, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> notices you as you enter!`, mob.Character.Name), true)
					cmdQueue.QueueCommand(0, mob.InstanceId, `lookfortrouble`, 4)
				}

			}

			response.NextCommand = "look secretly" // Force them to look at the new room they are in.
		}

	}

	if !response.Handled {

		if rest == "north" || rest == "south" || rest == "east" || rest == "west" || rest == "up" || rest == "down" || rest == "northwest" || rest == "northeast" || rest == "southwest" || rest == "southeast" {
			response.SendUserMessage(userId, "You're bumping into walls.", true)
			if !user.Character.HasBuffFlag(buffs.Hidden) {
				response.SendRoomMessage(room.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> is bumping into walls.`, user.Character.Name),
					true)
			}
			response.Handled = true
		}

	}

	return response, nil
}
