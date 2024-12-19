package usercommands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/races"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/users"
)

func Mob(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if user.Permission != users.PermissionAdmin {
		user.SendText(`<ansi fg="alert-4">Only admins can use this command</ansi>`)
		return true, nil
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {
		infoOutput, _ := templates.Process("admincommands/help/command.mob", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	// mob create
	if args[0] == `create` {

		return mob_Create(rest, user, room)

	}

	return true, nil
}

func mob_Create(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`mob`, rest)

	mobCreateAnswerName := ``
	mobCreateAnswerRace := 0
	mobCreateAnswerZone := ``
	mobCreateAnswerDescription := ``

	if isNew {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`Lets get a little info first.%s`, term.CRLFStr))
	}

	//
	// Name Selection
	//

	question := cmdPrompt.Ask(`What will the mob be called?`, []string{})
	if !question.Done {
		return true, nil
	}

	if question.Response == `` {
		user.SendText("Aborting...")
		user.ClearPrompt()
		return true, nil
	}

	mobCreateAnswerName = question.Response

	//
	// Race Selection
	//
	raceOptions := append([]races.Race{}, races.GetRaces()...)

	sort.SliceStable(raceOptions, func(i, j int) bool {
		return raceOptions[i].Name < raceOptions[j].Name
	})

	question = cmdPrompt.Ask(`What race will the mob be?`, []string{})
	if !question.Done {
		tplTxt, _ := templates.Process("character/start.racelist", raceOptions)
		user.SendText(tplTxt)
		return true, nil
	}

	if question.Response == `` {
		user.SendText("Aborting...")
		user.ClearPrompt()
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
		return Help(helpCmd+` `+helpRest, user, room)
	}

	raceNameSelection := question.Response
	if restNum, err := strconv.Atoi(raceNameSelection); err == nil {
		if restNum > 0 && restNum <= len(raceOptions) {
			raceNameSelection = raceOptions[restNum-1].Name
		}
	}

	for _, r := range races.GetRaces() {
		if strings.EqualFold(r.Name, raceNameSelection) {
			mobCreateAnswerRace = r.RaceId
		}
	}

	if mobCreateAnswerRace == 0 {
		question.RejectResponse()

		tplTxt, _ := templates.Process("character/start.racelist", raceOptions)
		user.SendText(tplTxt)

		return true, nil
	}

	//
	// Zone Selection
	//
	zoneOptions := append([]string{}, rooms.GetAllZoneNames()...)

	sort.SliceStable(zoneOptions, func(i, j int) bool {
		return zoneOptions[i] < zoneOptions[j]
	})

	question = cmdPrompt.Ask(`What zone is this mob from?`, []string{})
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list", zoneOptions)
		user.SendText(tplTxt)
		return true, nil
	}

	if question.Response == `` {
		user.SendText("Aborting...")
		user.ClearPrompt()
		return true, nil
	}

	zoneNameSelection := question.Response
	if restNum, err := strconv.Atoi(zoneNameSelection); err == nil {
		if restNum > 0 && restNum <= len(zoneOptions) {
			zoneNameSelection = zoneOptions[restNum-1]
		}
	}

	for _, z := range zoneOptions {
		if strings.EqualFold(z, zoneNameSelection) {
			mobCreateAnswerZone = z
		}
	}

	if mobCreateAnswerZone == `` {
		question.RejectResponse()

		tplTxt, _ := templates.Process("tables/numbered-list", zoneOptions)
		user.SendText(tplTxt)
		return true, nil
	}

	//
	// Description
	//
	question = cmdPrompt.Ask(`Enter a description for the mob:`, []string{})
	if !question.Done {
		return true, nil
	}

	mobCreateAnswerDescription = question.Response

	user.ClearPrompt()

	mobId, err := mobs.CreateNewMobFile(mobCreateAnswerName, mobCreateAnswerRace, mobCreateAnswerZone, mobCreateAnswerDescription)

	if err != nil {
		user.SendText("Error: " + err.Error())
		return true, nil
	}

	mobInst := mobs.GetMobSpec(mobId)

	user.SendText(``)
	user.SendText(`<ansi bg="red" fg="white-bold">MOB CREATED</ansi>`)
	user.SendText(``)
	user.SendText(`<ansi fg="yellow-bold">Mob Name:</ansi>  <ansi fg="white-bold">` + mobCreateAnswerName + `</ansi>`)
	user.SendText(`<ansi fg="yellow-bold">Mob Race:</ansi>  <ansi fg="white-bold">` + strconv.Itoa(mobCreateAnswerRace) + ` (` + raceNameSelection + `)</ansi>`)
	user.SendText(`<ansi fg="yellow-bold">Mob Zone:</ansi>  <ansi fg="white-bold">` + mobCreateAnswerZone + `</ansi>`)
	user.SendText(`<ansi fg="yellow-bold">Mob Desc:</ansi>  <ansi fg="white-bold">` + mobCreateAnswerDescription + `</ansi>`)
	user.SendText(`<ansi fg="yellow-bold">File Path:</ansi> <ansi fg="white-bold">` + mobInst.Filepath() + `</ansi>`)
	user.SendText(``)
	user.SendText(`<ansi fg="black-bold">note: Try <ansi fg="command">spawn mob ` + mobCreateAnswerName + `</ansi> to test it.`)

	return true, nil
}
