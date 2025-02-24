package usercommands

import (
	"strings"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobcommands"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Command(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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
		return true, nil
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

	return true, nil
}
