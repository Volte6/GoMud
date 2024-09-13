package usercommands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Start(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.RoomId != -1 {
		return response, errors.New(`only allowed in the void`)
	}

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`start`, rest)

	if isNew {
		response.SendUserMessage(userId, fmt.Sprintf(`You'll need to answer some questions.%s`, term.CRLFStr))
	}

	if user.Character.RaceId == 0 {

		raceOptions := []string{}
		for _, r := range races.GetRaces() {
			if r.Selectable {
				raceOptions = append(raceOptions, r.Name)
			}
		}
		raceOptions = append(raceOptions, `?`)

		question := cmdPrompt.Ask(`Which race will you be?`, raceOptions, `?`)
		if !question.Done {
			response.Handled = true
			return response, nil
		}

		if question.Response == `?` {

			question.RejectResponse()
			return Help(`races`, userId)

		}

		for _, r := range races.GetRaces() {
			if strings.EqualFold(r.Name, question.Response) {

				if r.Selectable {
					user.Character.RaceId = r.Id()
					user.Character.Validate()

					response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="magenta">Your ghostly form materializes into that of a %s!</ansi>%s`, r.Name, term.CRLFStr))
				}

			}
		}

	}

	if strings.EqualFold(user.Character.Name, user.Username) || len(user.Character.Name) == 0 {

		question := cmdPrompt.Ask(`What will you be known as (name)?`, []string{})
		if !question.Done {
			response.Handled = true
			return response, nil
		}

		if strings.EqualFold(question.Response, user.Username) {
			response.SendUserMessage(userId, `Your username cannot match your character name!`)
			question.RejectResponse()
			response.Handled = true
			return response, nil
		}

		if err := user.SetCharacterName(question.Response); err != nil {
			response.SendUserMessage(userId, err.Error())
			question.RejectResponse()
			response.Handled = true
			return response, nil
		}

		user.ClearPrompt()

		response.SendUserMessage(userId, fmt.Sprintf(`You will be known as <ansi fg="yellow-bold">%s</ansi>!%s`, user.Character.Name, term.CRLFStr))
	}

	response.SendUserMessage(userId, fmt.Sprintf(`<ansi fg="magenta">Suddenly, a vortex appears before you, drawing you in before you have any chance to react!</ansi>%s`, term.CRLFStr))

	rooms.MoveToRoom(user.UserId, 1)
	response.SendUserMessage(userId, `Welcome to Frostfang. You can <ansi fg="command">look</ansi> at the <ansi fg="itemname">sign</ansi> here!`)

	response.Handled = true
	return response, nil
}
