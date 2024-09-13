package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Unlock(rest string, userId int) (util.MessageQueue, error) {

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

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		response.SendUserMessage(userId, "Unlock what?")
		response.Handled = true
		return response, nil
	}

	containerName := room.FindContainerByName(args[0])
	exitName, exitRoomId := room.FindExitByName(args[0])

	if containerName != `` {

		container := room.Containers[containerName]

		if !container.Lock.IsLocked() {
			response.SendUserMessage(userId, "That's not locked.")
			response.Handled = true
			return response, nil
		}

		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, containerName)
		hasKey, _ := user.Character.HasKey(lockId, int(container.Lock.Difficulty))

		var backpackKeyItm items.Item = items.Item{}
		var hasBackpackKey bool = false
		if !hasKey {
			backpackKeyItm, hasBackpackKey = user.Character.FindKeyInBackpack(lockId)
		}

		if hasKey {
			container.Lock.SetUnlocked()
			room.Containers[containerName] = container

			response.SendUserMessage(userId, fmt.Sprintf(`You use a key to unlock the <ansi fg="container">%s</ansi>.`, containerName))
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="container">%s</ansi>.`, user.Character.Name, containerName))
		} else if hasBackpackKey {

			itmSpec := backpackKeyItm.GetSpec()

			container.Lock.SetUnlocked()
			room.Containers[containerName] = container

			// Key entries look like:
			// "key-<roomid>-<exitname>": "<itemid>"
			user.Character.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, backpackKeyItm.ItemId))
			user.Character.RemoveItem(backpackKeyItm)

			response.SendUserMessage(userId, fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to lock the <ansi fg="container">%s</ansi>, and add it to your key ring for the future.`, itmSpec.Name, containerName))
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to lock the <ansi fg="container">%s</ansi>.`, user.Character.Name, containerName),
				userId)
		} else {
			response.SendUserMessage(userId, `You do not have the key for that. Maybe you could <ansi fg="command">picklock</ansi> the lock.`)
		}

		response.Handled = true
		return response, nil

	} else if exitRoomId > 0 {

		exitInfo := room.Exits[exitName]

		if !exitInfo.Lock.IsLocked() {
			response.SendUserMessage(userId, "That's not locked.")
			response.Handled = true
			return response, nil
		}

		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, exitName)
		hasKey, _ := user.Character.HasKey(lockId, int(exitInfo.Lock.Difficulty))

		var backpackKeyItm items.Item = items.Item{}
		var hasBackpackKey bool = false
		if !hasKey {
			backpackKeyItm, hasBackpackKey = user.Character.FindKeyInBackpack(lockId)
		}

		if hasKey {
			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo

			response.SendUserMessage(userId, fmt.Sprintf(`You use a key to unlock the <ansi fg="exit">%s</ansi> lock.`, exitName))
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName))
		} else if hasBackpackKey {

			itmSpec := backpackKeyItm.GetSpec()

			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo

			// Key entries look like:
			// "key-<roomid>-<exitname>": "<itemid>"
			user.Character.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, backpackKeyItm.ItemId))
			user.Character.RemoveItem(backpackKeyItm)

			response.SendUserMessage(userId, fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to unlock the <ansi fg="exit">%s</ansi> exit, and add it to your key ring for the future.`, itmSpec.Name, exitName))
			response.SendRoomMessage(room.RoomId,
				fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName),
				userId)
		} else {
			response.SendUserMessage(userId, `You do not have the key for that. Maybe you could <ansi fg="command">picklock</ansi> the lock.`)
		}

		response.Handled = true
		return response, nil

	}

	response.SendUserMessage(userId, "There is no such exit or container.")
	response.Handled = true
	return response, nil

}
