package usercommands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func SetRace(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.RoomId != -1 {
		return response, errors.New(`Only allowed in the viod`)
	}

	if rest == `` {
		return Help(`setrace`, userId, cmdQueue)
	}

	for _, r := range races.GetRaces() {
		if strings.EqualFold(r.Name, rest) {

			if r.Selectable {
				user.Character.RaceId = r.Id()
				user.Character.Validate()

				response.SendUserMessage(userId, `Race set to `+r.Name, true)

				response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="magenta">Your ghostly form materializes into that of a %s!</ansi>`, r.Name), true)
				response.SendUserMessage(userId, `<ansi fg="magenta">Suddenly, a vortex appears before you, drawing you in before you have any chance to react!</ansi>`, true)

				rooms.MoveToRoom(user.UserId, 1)
				response.SendUserMessage(userId, `Welcome to Frostfang. You can <ansi fg="command">look</ansi> at the <ansi fg="itemname">sign</ansi> here!`, true)
				response.Handled = true
				return response, nil
			}

			response.SendUserMessage(userId, `Only humans are allowed at this time.`, true)
			break
		}
	}

	response.SendUserMessage(userId, `Only humans are allowed at this time.`, true)

	response.Handled = true
	return response, nil
}
