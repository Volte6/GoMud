package usercommands

import (
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Go(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Aggro != nil {
		user.SendText("You can't do that! You are in combat!")
		return true, nil
	}

	// If has a buff that prevents combat, skip the player
	if user.Character.HasBuffFlag(buffs.NoMovement) {
		user.SendText("You can't do that!")
		return true, nil
	}

	isSneaking := user.Character.HasBuffFlag(buffs.Hidden)

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	handled := false

	exitName, goRoomId := room.FindExitByName(rest)

	if goRoomId > 0 || exitName != `` {

		actionCost := 10
		encumbered := false
		if len(user.Character.Items) > user.Character.CarryCapacity() {
			actionCost = 50
			encumbered = true
		}

		if !user.Character.DeductActionPoints(actionCost) {

			if encumbered {
				user.SendText("You're too tired to move!")
			} else {
				user.SendText("You're too encumbered to move!")
			}

			return true, nil
		}

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

				user.SendText(`You know this lock well, you quickly pick it.`)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> quickly picks the lock on the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
					userId)

				exitInfo.Lock.SetUnlocked()
				room.Exits[exitName] = exitInfo

			} else if hasKey {
				user.SendText(fmt.Sprintf(`You use the key on your key ring to unlock the <ansi fg="exit">%s</ansi> exit.`, exitName))
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
					userId)

				exitInfo.Lock.SetUnlocked()
				room.Exits[exitName] = exitInfo
			} else {

				// check for a key item on their person
				if backpackKeyItm, hasBackpackKey := user.Character.FindKeyInBackpack(lockId); hasBackpackKey {

					itmSpec := backpackKeyItm.GetSpec()

					user.SendText(fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to unlock the <ansi fg="exit">%s</ansi> exit, and add it to your key ring for the future.`, itmSpec.Name, exitName))
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
						userId)

					// Key entries look like:
					// "key-<roomid>-<exitname>": "<itemid>"
					user.Character.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, backpackKeyItm.ItemId))
					user.Character.RemoveItem(backpackKeyItm)

					exitInfo.Lock.SetUnlocked()
					room.Exits[exitName] = exitInfo

				}

				if exitInfo.Lock.IsLocked() {
					user.SendText(`There's a lock preventing you from going that way. You'll need a <ansi fg="item">Key</ansi> or to <ansi fg="command">pick</ansi> the lock with <ansi fg="item">lockpicks</ansi>.`)
					return true, nil
				}
			}

		}

		originRoomId := user.Character.RoomId

		// Load current room details
		destRoom := rooms.LoadRoom(goRoomId)
		if destRoom == nil {
			return false, fmt.Errorf(`room %d not found`, goRoomId)
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
			user.SendText("Oops, couldn't move there!")
		} else {

			scripting.TryRoomScriptEvent(`onExit`, user.UserId, originRoomId)

			c := configs.GetConfig()

			// Tell the player they are moving
			if isSneaking {
				user.SendText(
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`You <ansi fg="black-bold">sneak</ansi> towards the %s exit.`, exitName),
					))
			} else {
				user.SendText(
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`You head towards the <ansi fg="exit">%s</ansi> exit.`, exitName),
					))

				// Tell the old room they are leaving
				room.SendText(
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`<ansi fg="username">%s</ansi> leaves towards the <ansi fg="exit">%s</ansi> exit.`, user.Character.Name, exitName),
					),
					userId)
				// Tell the new room they have arrived
				destRoom.SendText(
					fmt.Sprintf(string(c.EnterRoomMessageWrapper),
						fmt.Sprintf(`<ansi fg="username">%s</ansi> enters from %s.`, user.Character.Name, enterFromExit),
					),
					userId)

				destRoom.SendTextToExits(`You hear someone moving around.`, true, room.GetPlayers(rooms.FindAll)...)
			}

			if currentParty := parties.Get(userId); currentParty != nil {

				if currentParty.IsLeader(userId) {

					for _, partyMemberId := range currentParty.UserIds {
						if partyMemberId == userId {
							continue
						}
						if partyUser := users.GetByUserId(partyMemberId); partyUser != nil {
							if partyUser.Character.RoomId == room.RoomId {
								partyUser.SendText(`You follow the party leader.`)
								partyUser.Command(rest)
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

					mob.Command(rest)

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

					speedDelta := mob.Character.Stats.Speed.ValueAdj - user.Character.Stats.Speed.ValueAdj
					if speedDelta < 1 {
						speedDelta = 1
					}

					// Chance that a mob follows the player
					targetVal := 20 + mob.Character.Stats.Perception.ValueAdj + speedDelta

					roll := util.Rand(100)

					util.LogRoll(`Mob Follow`, roll, targetVal)

					if roll >= targetVal {
						continue
					}

					mob.Command(rest)

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

					user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> notices you as you enter!`, mob.Character.Name))

					mob.Command(`lookfortrouble`, 4)

				}

			}

			handled = true
			Look(`secretly`, userId)

			scripting.TryRoomScriptEvent(`onEnter`, user.UserId, destRoom.RoomId)
		}

	}

	if !handled {

		if rest == "north" || rest == "south" || rest == "east" || rest == "west" || rest == "up" || rest == "down" || rest == "northwest" || rest == "northeast" || rest == "southwest" || rest == "southeast" {
			user.SendText("You're bumping into walls.")
			if !user.Character.HasBuffFlag(buffs.Hidden) {

				c := configs.GetConfig()

				room.SendText(
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`<ansi fg="username">%s</ansi> is bumping into walls.`, user.Character.Name),
					),
					userId)
			}
			handled = true
		}

	}

	return handled, nil
}
