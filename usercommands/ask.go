package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"

	"github.com/volte6/mud/users"
)

func Ask(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	// Core "useful" commands
	usefulCommands := []string{
		`attack`,
		`give`,
		`get`,
		`drop`,
		`equip`,
		`remove`,
	}

	// Additional commands that are more for fun
	allowedCommands := []string{
		`say`,
		`look`,
		`emote`,
		`throw`,
	}

	// args should look like one of the following:
	// target buffId - put buff on target if in the room
	// buffId - put buff on self
	// search searchTerm - search for buff by name, display results
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {

		for _, mId := range room.GetMobs(rooms.FindCharmed) {
			mob := mobs.GetInstance(mId)
			if mob == nil {
				continue
			}
			if mob.Character.IsCharmed(userId) {

				cmdQueue.QueueCommand(0, mId, fmt.Sprintf(`say I can do a few useful things, such as %s`,
					fmt.Sprintf(`<ansi fg="command">%s</ansi>`, strings.Join(usefulCommands, `</ansi>, <ansi fg="command">`)),
				))
				cmdQueue.QueueCommand(0, mId, fmt.Sprintf(`say I can do some other stuff, like %s`,
					fmt.Sprintf(`<ansi fg="command">%s</ansi>`, strings.Join(allowedCommands, `</ansi>, <ansi fg="command">`)),
				))

				response.Handled = true
				return response, nil
			}
		}

		response.SendUserMessage(userId, `You must <ansi fg="command">ask</ansi> <ansi fg="mobname">someone</ansi> <ansi fg="yellow">something</ansi>`, true)
		response.Handled = true
		return response, nil
	}

	allowedCommands = append(allowedCommands, usefulCommands...)

	searchName := args[0]

	// Only ask charmed players or mobs to do stuff
	_, mobId := room.FindByName(searchName)

	if mobId > 0 {

		mob := mobs.GetInstance(mobId)
		if mob == nil {
			response.SendUserMessage(userId, `Nobody found by that name`, true)
			response.Handled = true
			return response, nil
		}

		args = args[1:]

		if !mob.Character.IsCharmed() {
			response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> asks <ansi fg="mobname">%s</ansi> about "%s"`, user.Character.Name, mob.Character.Name, strings.Join(args, ` `)), true, user.UserId)
		}

		// players may type "ask <mob> to <do something>"
		if len(args) > 1 && strings.ToLower(args[0]) == `to` {
			args = args[1:]
		}
		if len(args) > 1 && strings.ToLower(args[0]) == `about` {
			args = args[1:]
		}

		cmd := args[0]
		args = args[1:]

		rest = strings.Join(args, ` `)

		if mob.Character.IsCharmed(userId) {
			// If an alias was entered, conovert it
			if newCmd, ok := aliases[cmd]; ok {
				cmd = newCmd
			}

			if cmd == `attack` {
				if pid, _ := room.FindByName(rest); pid > 0 {
					if !configs.GetConfig().PVPEnabled {
						cmdQueue.QueueCommand(0, mobId, `emote shakes their head.`)
						cmdQueue.QueueCommand(0, mobId, `say PVP is currently disabled.`)
						response.Handled = true
						return response, nil
					}
				}
			}

			// Check if actual command is allowed
			for _, allowedCmd := range allowedCommands {
				if cmd == allowedCmd {
					cmdQueue.QueueCommand(0, mobId, fmt.Sprintf(`%s %s`, cmd, rest))

					response.Handled = true
					return response, nil
				}
			}
		}

		searchText := strings.TrimSpace(fmt.Sprintf(`%s %s`, cmd, rest))
		for _, subject := range mob.AskSubjects {

			// They must have this quest, or continue
			if len(subject.IfQuest) > 0 && !user.Character.HasQuest(subject.IfQuest) {
				continue
			}

			// They must not have this quest, or continue
			if len(subject.IfNotQuest) > 0 && user.Character.HasQuest(subject.IfNotQuest) {
				continue
			}

			for _, noun := range subject.AskNouns {

				if noun == searchText || strings.Contains(searchText, noun) {
					for _, reply := range subject.ReplyCommands {
						reply = strings.ReplaceAll(reply, `{userid}`, fmt.Sprintf(`@%d`, user.UserId)) // @ prefix denotes a userId
						cmdQueue.QueueCommand(0, mobId, reply)
					}
					response.Handled = true
					return response, nil
				}
			}
		}

		cmdQueue.QueueCommand(0, mobId, `emote shakes their head.`)
	}

	response.Handled = true
	return response, nil
}
