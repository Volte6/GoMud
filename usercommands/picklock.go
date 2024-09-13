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

func Picklock(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	lockpickItm := items.Item{}
	for _, itm := range user.Character.GetAllBackpackItems() {
		if itm.GetSpec().Type == items.Lockpicks {
			lockpickItm = itm
			break
		}
	}

	if lockpickItm.ItemId < 1 {
		user.SendText(`You need <ansi fg="item">lockpicks</ansi> to pick a lock.`)
		return true, ``, nil
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		user.SendText("You wanna pock a lock? Specify where it is.")
		return true, ``, nil
	}

	lockId := ``
	lockStrength := 0

	containerName := room.FindContainerByName(args[0])
	exitName, exitRoomId := room.FindExitByName(args[0])

	if containerName != `` {

		container := room.Containers[containerName]

		if !container.HasLock() {
			user.SendText("There is no lock there.")
			return true, ``, nil
		}

		if !container.Lock.IsLocked() {
			user.SendText("It's already unlocked.")
			return true, ``, nil
		}

		args = args[1:]
		lockStrength = int(container.Lock.Difficulty)
		lockId = fmt.Sprintf(`%d-%s`, room.RoomId, containerName)

	} else if exitRoomId > 0 {

		// get the first entry int he slice and shorten the slice
		args = args[1:]

		exitInfo := room.Exits[exitName]

		if !exitInfo.HasLock() {
			user.SendText("There is no lock there.")
			return true, ``, nil
		}

		if !exitInfo.Lock.IsLocked() {
			user.SendText("It's already unlocked.")
			return true, ``, nil
		}

		lockStrength = int(exitInfo.Lock.Difficulty)
		lockId = fmt.Sprintf(`%d-%s`, room.RoomId, exitName)

	} else {

		user.SendText("There is no such exit or container.")
		return true, ``, nil
	}

	//
	// Most of what follows shouldn't reference an exit or a chest, but rather lock details.
	//
	keyring_sequence := user.Character.GetKey(lockId)

	sequence := util.GetLockSequence(lockId, lockStrength, string(configs.GetConfig().Seed))

	if keyring_sequence == sequence {
		user.SendText("")
		user.SendText("Your keyring already has this lock on it.")

		user.ClearPrompt()

		user.SendText(``)
		user.SendText(`<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You Successfully picked the lock!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		user.SendText(`<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You can automatically pick this lock any time as long as you carry <ansi fg="item">lockpicks</ansi>!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		user.SendText(``)

		if containerName != `` {

			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="container">%s</ansi> lock`, user.Character.Name, containerName), userId)

			container := room.Containers[containerName]
			container.Lock.SetUnlocked()
			room.Containers[containerName] = container
		} else {

			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName), userId)

			exitInfo := room.Exits[exitName]
			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo
		}
		return true, ``, nil
	}

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`picklock`, rest)

	if isNew {
		user.SendText(GetLockRender(sequence, keyring_sequence))
	}

	entered := ``
	if len(keyring_sequence) > 0 {
		entered = keyring_sequence
	}

	question := cmdPrompt.Ask(`Move your lockpick?`, []string{`UP`, `DOWN`, `quit`})
	if !question.Done {
		return true, ``, nil
	}

	if question.Response == `quit` {
		user.ClearPrompt()
		user.SendText(`Type '<ansi fg="command">help picklock</ansi>' for more information on picking locks.`)
		return true, ``, nil
	}

	direction := question.Response

	question.RejectResponse() // Always reset this question, since we want to keep reusing it.

	r := strings.ToUpper(direction)
	r = string(r[0])

	if r != "U" && r != "D" {
		return true, ``, nil
	}

	entered += r

	for i := 0; i < len(entered); i++ {
		if entered[i] != sequence[i] {
			// Mismatch! BREAKS!
			entered = ``
			user.Character.UseItem(lockpickItm)

			user.SendText(``)
			user.SendText(fmt.Sprintf(`<ansi fg="yellow-bold">***</ansi> <ansi fg="red-bold">Oops! Your <ansi fg="item">%s</ansi> break off in the lock, resetting the lock. You'll have to start all over.</ansi> <ansi fg="yellow-bold">***</ansi>`, lockpickItm.GetSpec().NameSimple))
			user.SendText(``)
		}
	}

	user.Character.SetKey(lockId, entered)

	if len(entered) > 0 {
		user.SendText(``)
		user.SendText(`<ansi fg="green-bold">A satisfying *click* tells you that you're making progress...</ansi>`)
	} else {
		user.ClearPrompt()
		return true, ``, nil
	}

	user.SendText(GetLockRender(sequence, entered))

	if sequence == entered {

		user.SendText(``)
		user.SendText(`<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You Successfully picked the lock!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		user.SendText(`<ansi fg="yellow-bold">***</ansi> <ansi fg="green-bold">You can automatically pick this lock any time as long as you carry <ansi fg="item">lockpicks</ansi>!</ansi> <ansi fg="yellow-bold">***</ansi>`)
		user.SendText(``)

		if containerName != `` {

			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="container">%s</ansi> lock`, user.Character.Name, containerName), userId)

			container := room.Containers[containerName]
			container.Lock.SetUnlocked()
			room.Containers[containerName] = container
		} else {

			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> picks the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName), userId)

			exitInfo := room.Exits[exitName]
			exitInfo.Lock.SetUnlocked()
			room.Exits[exitName] = exitInfo
		}

		user.ClearPrompt()

		return true, ``, nil

	} else {
		if containerName != `` {
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to pick the <ansi fg="container">%s</ansi> lock`, user.Character.Name, containerName), userId)
		} else {
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> tries to pick the <ansi fg="exit">%s</ansi> lock`, user.Character.Name, exitName), userId)
		}
	}

	return true, ``, nil
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
