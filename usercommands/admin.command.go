package usercommands

import (
	"fmt"
	"strings"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/mobcommands"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Command(rest string, userId int) (bool, string, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, ``, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, ``, fmt.Errorf(`room %d not found`, user.Character.RoomId)
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
		user.SendText(infoOutput)
		return true, ``, nil
	}

	searchName := args[0]
	args = args[1:]
	cmd := strings.TrimPrefix(rest, searchName+` `)
	//cmd := strings.Join(args, ` `)

	playerId, mobId := room.FindByName(searchName)

	// Use the index for how many turns to defer the extra commands
	for waitTurns, oneCmd := range strings.Split(cmd, `;`) {
		if mobId > 0 {

			events.AddToQueue(events.Input{
				MobInstanceId: mobId,
				InputText:     oneCmd,
				WaitTurns:     waitTurns,
			})

		} else if playerId > 0 {

			events.AddToQueue(events.Input{
				UserId:    playerId,
				InputText: oneCmd,
				WaitTurns: waitTurns,
			})

		}
	}

	return true, ``, nil
}
