package usercommands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
)

func Start(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Character.RoomId != -1 {
		return false, errors.New(`only allowed in the void`)
	}

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`start`, rest)

	if isNew {
		user.SendText(fmt.Sprintf(`You'll need to answer some questions.%s`, term.CRLFStr))
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
			return true, nil
		}

		if question.Response == `?` {

			question.RejectResponse()
			return Help(`races`, user, room)

		}

		for _, r := range races.GetRaces() {
			if strings.EqualFold(r.Name, question.Response) {

				if r.Selectable {
					user.Character.RaceId = r.Id()
					user.Character.Alignment = r.DefaultAlignment
					user.Character.Validate()

					user.SendText(fmt.Sprintf(`<ansi fg="magenta">Your ghostly form materializes into that of a %s!</ansi>%s`, r.Name, term.CRLFStr))
				}

			}
		}

	}

	if strings.EqualFold(user.Character.Name, user.Username) || len(user.Character.Name) == 0 || strings.ToLower(user.Character.Name) == `nameless` {

		question := cmdPrompt.Ask(`What will you be known as (name)?`, []string{})
		if !question.Done {
			return true, nil
		}

		if strings.EqualFold(question.Response, user.Username) {
			user.SendText(`Your username cannot match your character name!`)
			question.RejectResponse()
			return true, nil
		}

		for _, c := range characters.LoadAlts(user.Username) {
			if strings.EqualFold(question.Response, c.Name) {
				user.SendText(`Your already have a character named that!`)
				question.RejectResponse()
				return true, nil
			}
		}

		if err := user.SetCharacterName(question.Response); err != nil {
			user.SendText(err.Error())
			question.RejectResponse()
			return true, nil
		}

		user.ClearPrompt()

		user.SendText(fmt.Sprintf(`You will be known as <ansi fg="yellow-bold">%s</ansi>!%s`, user.Character.Name, term.CRLFStr))
	}

	user.SendText(fmt.Sprintf(`<ansi fg="magenta">Suddenly, a vortex appears before you, drawing you in before you have any chance to react!</ansi>%s`, term.CRLFStr))

	for _, ridStr := range configs.GetConfig().TutorialStartRooms {

		rid, _ := strconv.ParseInt(ridStr, 10, 64)
		skip := false

		for _, populatedRoomId := range rooms.GetRoomsWithPlayers() {
			roomCt := 10
			for i := 0; i < roomCt; i++ {
				if int(rid)+i == populatedRoomId {
					skip = true
					continue
				}
			}
		}

		if skip {
			continue
		}

		if _, err := scripting.TryRoomScriptEvent(`onEnter`, user.UserId, int(rid)); err == nil {

			rooms.MoveToRoom(user.UserId, int(rid))
			return true, nil
		}

	}

	user.SendText(`Someone else is currently utilizing the tutorial, please try again in a few minutes.`)

	//rooms.MoveToRoom(user.UserId, 1)
	//user.SendText(`Welcome to Frostfang. You can <ansi fg="command">look</ansi> at the <ansi fg="itemname">sign</ansi> here!`)

	return true, nil
}
