package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/mobcommands"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Command(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

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

	// args should look like one of the following:
	// target buffId - put buff on target if in the room
	// buffId - put buff on self
	// search searchTerm - search for buff by name, display results
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		mobCommands := mobcommands.GetAllMobCommands()

		infoOutput, _ := templates.Process("admincommands/help/command.command", mobCommands)
		response.Handled = true
		response.SendUserMessage(userId, infoOutput, false)
		return response, nil
	}

	searchName := args[0]
	args = args[1:]
	cmd := strings.TrimPrefix(rest, searchName+` `)
	//cmd := strings.Join(args, ` `)

	playerId, mobId := room.FindByName(searchName)

	// Use the index for how many turns to defer the extra commands
	for waitTurns, oneCmd := range strings.Split(cmd, `;`) {
		if mobId > 0 {
			cmdQueue.QueueCommand(0, mobId, oneCmd, waitTurns)
		} else if playerId > 0 {
			cmdQueue.QueueCommand(playerId, 0, oneCmd, waitTurns)
		}
	}

	response.Handled = true
	return response, nil
}
