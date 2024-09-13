package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Picklock(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	lockpickItm := items.Item{}
	for _, itm := range user.Character.GetAllBackpackItems() {
		if itm.GetSpec().Type == items.Lockpicks {
			lockpickItm = itm
			break
		}
	}

	if lockpickItm.ItemId < 1 {
		response.SendUserMessage(userId, `You need <ansi fg="item">lockpicks</ansi> to pick a lock.`)
		response.Handled = true
		return response, nil
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		response.SendUserMessage(userId, "You wanna pock a lock? Specify where it is.")
		response.Handled = true
		return response, nil
	}

	lockId := ``
	lockStrength := 0

	containerName := room.FindContainerByName(args[0])
	exitName, exitRoomId := room.FindExitByName(args[0])

	if containerName != `` {

		container := room.Containers[containerName]

		if !container.HasLock() {
			response.SendUserMessage(userId, "There is no lock there.")
			response.Handled = true
			return response, nil
		}

		if !container.Lock.IsLocked() {
			response.SendUserMessage(userId, "It's already unlocked.")
			response.Handled = true
			return response, nil
		}

		args = args[1:]
		lockStrength = int(container.Lock.Difficulty)
		lockId = fmt.Sprintf(`%d-%s`, room.RoomId, containerName)

	} else if exitRoomId > 0 {

		// get the first entry int he slice and shorten the slice
		args = args[1:]

		exitInfo := room.Exits[exitName]

		if !exitInfo.HasLock() {
			response.SendUserMessage(userId, "There is no lock there.")
			response.Handled = true
			return response, nil
		}

		if !exitInfo.Lock.IsLocked() {
			response.SendUserMessage(userId, "It's already unlocked.")
			response.Handled = true
			return response, nil
		}

		lockStrength = int(exitInfo.Lock.Difficulty)
		lockId = fmt.Sprintf(`%d-%s`, room.RoomId, exitName)

	} else {

		response.SendUserMessage(userId, "There is no such exit or container.")
		response.Handled = true
		return response, nil
	}

	//
	// Most of what follows shouldn't reference an exit or a chest, but rather lock details.
	//
	keyring_sequence := user.Character.GetKey(lockId)

	sequence := util.GetLockSequence(lockId, lockStrength, string(configs.GetConfig().Seed))

	if keyring_sequence == sequence {
		response.SendUserMessage(userId, "")
		response.SendUserMessage(userId, "Your keyring already has this lock on it.")

		user.ClearPrompt()

		response.SendUserMessage(userId, ``)
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You Successfully picked the lock!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You can automatically pick this lock any time as long as you carry <ansi fg="item">lockpicks</ansi>!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		response.SendUserMessage(userId, ``)

		if containerName != `` {

			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="container">%s</ansi> lock`, user.Character.Name, containerName))

			container := room.Containers[containerName]
			container.Lock.SetUnlocked()
			room.Containers[containerName] = container
		} else {

			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName))

			exitInfo := room.Exits[exitName]
			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo
		}
		response.Handled = true
		return response, nil
	}

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`picklock`, rest)

	if isNew {
		response.SendUserMessage(userId, GetLockRender(sequence, keyring_sequence))
	}

	entered := ``
	if len(keyring_sequence) > 0 {
		entered = keyring_sequence
	}

	question := cmdPrompt.Ask(`Move your lockpick?`, []string{`UP`, `DOWN`, `quit`})
	if !question.Done {
		response.Handled = true
		return response, nil
	}

	if question.Response == `quit` {
		user.ClearPrompt()
		response.SendUserMessage(userId, `Type '<ansi fg="command">help picklock</ansi>' for more information on picking locks.`)
		response.Handled = true
		return response, nil
	}

	direction := question.Response

	question.RejectResponse() // Always reset this question, since we want to keep reusing it.

	r := strings.ToUpper(direction)
	r = string(r[0])

	if r != "U" && r != "D" {
		response.Handled = true
		return response, nil
	}

	entered += r

	for i := 0; i < len(entered); i++ {
		if entered[i] != sequence[i] {
			// Mismatch! BREAKS!
			entered = ``
			user.Character.UseItem(lockpickItm)

			response.SendUserMessage(userId, ``)
			response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="yellow-bold">***</ansi> <ansi fg="red-bold">Oops! Your <ansi fg="item">%s</ansi> break off in the lock, resetting the lock. You'll have to start all over.</ansi> <ansi fg="yellow-bold">***</ansi>`, lockpickItm.GetSpec().NameSimple))
			response.SendUserMessage(userId, ``)
		}
	}

	user.Character.SetKey(lockId, entered)

	if len(entered) > 0 {
		response.SendUserMessage(userId, ``)
		response.SendUserMessage(userId, `<ansi fg="green-bold">A satisfying *click* tells you that you're making progress...</ansi>`)
	} else {
		user.ClearPrompt()
		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId, GetLockRender(sequence, entered))

	if sequence == entered {

		response.SendUserMessage(userId, ``)
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You Successfully picked the lock!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		response.SendUserMessage(userId, `<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You can automatically pick this lock any time as long as you carry <ansi fg="item">lockpicks</ansi>!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		response.SendUserMessage(userId, ``)

		if containerName != `` {

			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="container">%s</ansi> lock`, user.Character.Name, containerName))

			container := room.Containers[containerName]
			container.Lock.SetUnlocked()
			room.Containers[containerName] = container
		} else {

			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName))

			exitInfo := room.Exits[exitName]
			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo
		}

		user.ClearPrompt()

		response.Handled = true
		return response, nil

	} else {
		if containerName != `` {
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to pick the <ansi fg="container">%s</ansi> lock`, user.Character.Name, containerName))
		} else {
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to pick the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName))
		}
	}

	response.Handled = true
	return response, nil
}

func GetLockRender(sequence string, entered string) string {

	rows := [][]string{}

	if len(entered) > len(sequence) {
		entered = entered[:len(sequence)]
	}

	formatting := make([]string, len(sequence))

	row := []string{}
	for i := 0; i < len(sequence); i++ {
		if len(entered) > i && entered[i] == sequence[i] && entered[i] == 'U' {
			row = append(row, `  U  `)
			formatting[i] = `<ansi fg="green-bold">%s</ansi>`
		} else {
			row = append(row, `     `)
		}
	}
	rows = append(rows, row)

	row = []string{}
	for i := 0; i < len(sequence); i++ {
		if i >= len(entered) || entered[i] != sequence[i] {
			row = append(row, `  ?  `)
			formatting[i] = `<ansi fg="red-bold">%s</ansi>`
		} else {
			if entered[i] == 'U' {
				row = append(row, `  ↑  `)
			} else if entered[i] == 'D' {
				row = append(row, `  ↓  `)
			} else {
				row = append(row, `     `)
			}
		}
	}
	rows = append(rows, row)

	row = []string{}
	for i := 0; i < len(sequence); i++ {
		if len(entered) > i && entered[i] == sequence[i] && entered[i] == 'D' {
			row = append(row, `  D  `)
			formatting[i] = `<ansi fg="green-bold">%s</ansi>`
		} else {
			row = append(row, `     `)
		}
	}
	rows = append(rows, row)

	picklockTable := templates.GetTable(`The Lock Sequence Looks like:`, rows[0], rows, formatting)
	tplTxt, _ := templates.Process("tables/lockpicking", picklockTable)

	return tplTxt

}
