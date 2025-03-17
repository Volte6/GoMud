package usercommands

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/races"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

func Start(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user.Character.RoomId != -1 {
		return false, errors.New(`only allowed in the void`)
	}

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`start`, rest)

	if isNew {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`You'll need to answer some questions.%s`, term.CRLFStr))
	}

	if user.Character.RaceId == 0 {

		raceOptions := []templates.NameDescription{}

		for _, r := range races.GetRaces() {
			if r.Selectable {
				raceOptions = append(raceOptions, templates.NameDescription{
					Id:          r.RaceId,
					Name:        r.Name,
					Description: r.Description,
				})
			}
		}
		sort.SliceStable(raceOptions, func(i, j int) bool {
			return raceOptions[i].Name < raceOptions[j].Name
		})

		question := cmdPrompt.Ask(`Which race will you be?`, []string{})
		if !question.Done {

			tplTxt, _ := templates.Process("tables/numbered-list", raceOptions)
			user.SendText(tplTxt)
			user.SendText(`  Want to know more details? Type <ansi fg="command">help {racename}</ansi> or <ansi fg="command">help {number}</ansi>`)
			user.SendText(``)
			return true, nil
		}

		respLower := strings.ToLower(question.Response)
		if len(respLower) >= 5 && respLower[0:5] == `help ` {
			helpCmd := `race`
			helpRest := respLower[5:]

			if restNum, err := strconv.Atoi(helpRest); err == nil {
				if restNum > 0 && restNum <= len(raceOptions) {
					helpRest = raceOptions[restNum-1].Name
				} else {
					helpCmd = `races`
					helpRest = ``
				}
			}

			question.RejectResponse()
			return Help(helpCmd+` `+helpRest, user, room, flags)
		}

		raceNameSelection := question.Response
		if restNum, err := strconv.Atoi(raceNameSelection); err == nil {
			if restNum > 0 && restNum <= len(raceOptions) {
				raceNameSelection = raceOptions[restNum-1].Name
			}
		}

		matchFound := false
		for _, r := range races.GetRaces() {
			if strings.EqualFold(r.Name, raceNameSelection) {

				if r.Selectable {
					matchFound = true
					user.Character.RaceId = r.Id()
					user.Character.Alignment = r.DefaultAlignment
					user.Character.Validate()

					user.SendText(``)
					user.SendText(fmt.Sprintf(`  <ansi fg="magenta">*** Your ghostly form materializes into that of a %s ***</ansi>%s`, r.Name, term.CRLFStr))
					break
				}

			}
		}

		if !matchFound {
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-list", raceOptions)
			user.SendText(tplTxt)
			user.SendText(`  Want to know more details? Type <ansi fg="command">help {racename}</ansi> or <ansi fg="command">help {number}</ansi>`)
			user.SendText(``)

			return true, nil
		}
	}

	if strings.EqualFold(user.Character.Name, user.Username) || len(user.Character.Name) == 0 || strings.ToLower(user.Character.Name) == `nameless` {

		question := cmdPrompt.Ask(`What will your character be known as (name)?`, []string{})
		if !question.Done {
			return true, nil
		}

		if strings.EqualFold(question.Response, user.Username) {
			user.SendText(`Your username cannot match your character name!`)
			question.RejectResponse()
			return true, nil
		}

		for _, c := range characters.LoadAlts(user.UserId) {
			if strings.EqualFold(question.Response, c.Name) {
				user.SendText(`Your already have a character named that!`)
				question.RejectResponse()
				return true, nil
			}
		}

		if err := users.ValidateName(question.Response); err != nil {
			user.SendText(`that name is not allowed: ` + err.Error())
			question.RejectResponse()
			return true, nil
		}

		if bannedPattern, ok := configs.GetConfig().IsBannedName(question.Response); ok {
			user.SendText(`that username matched the prohibited name pattern: "` + bannedPattern + `"`)
			question.RejectResponse()
			return true, nil
		}

		if foundUserId, _ := users.CharacterNameSearch(question.Response); foundUserId > 0 {
			user.SendText(`that character name is already in use.`)
			question.RejectResponse()
			return true, nil
		}

		for _, name := range mobs.GetAllMobNames() {
			if strings.EqualFold(name, question.Response) {
				user.SendText("that name is in use")
				question.RejectResponse()
				return true, nil
			}
		}

		usernameSelected := question.Response

		question = cmdPrompt.Ask(`Choose the name <ansi fg="username">`+usernameSelected+`</ansi>?`, []string{`yes`, `no`}, `no`)
		if !question.Done {
			return true, nil
		}

		if question.Response == `no` {
			user.ClearPrompt()
			return Start(rest, user, room, flags)
		}

		if err := user.SetCharacterName(usernameSelected); err != nil {
			user.SendText(err.Error())
			question.RejectResponse()
			return true, nil
		}

		user.ClearPrompt()

		user.SendText(fmt.Sprintf(`You will be known as <ansi fg="yellow-bold">%s</ansi>!%s`, user.Character.Name, term.CRLFStr))
	}

	user.Character.ExtraLives = int(configs.GetGamePlayConfig().LivesStart)

	user.EventLog.Add(`char`, fmt.Sprintf(`Created a new character: <ansi fg="username">%s</ansi>`, user.Character.Name))

	user.SendText(fmt.Sprintf(`<ansi fg="magenta">Suddenly, a vortex appears before you, drawing you in before you have any chance to react!</ansi>%s`, term.CRLFStr))

	for _, ridStr := range configs.GetSpecialRoomsConfig().TutorialStartRooms {

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
			user.SetConfigOption(`tinymap`, true)
			rooms.MoveToRoom(user.UserId, int(rid))
			return true, nil
		}

	}

	user.SendText(`Someone else is currently utilizing the tutorial, please try again in a few minutes.`)

	//rooms.MoveToRoom(user.UserId, 1)
	//user.SendText(`Welcome to Frostfang. You can <ansi fg="command">look</ansi> at the <ansi fg="itemname">sign</ansi> here!`)

	return true, nil
}
