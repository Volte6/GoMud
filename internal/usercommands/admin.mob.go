package usercommands

import (
	"fmt"
	"log/slog"
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

	if args[0] == `spawn` {
		return mob_Spawn(strings.TrimSpace(rest[5:]), user, room)
	}

	return true, nil
}

func mob_Spawn(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	// special handling of loot goblin
	if rest == `loot goblin` {
		if gRoom := rooms.LoadRoom(rooms.GoblinRoom); gRoom != nil { // loot goblin room
			user.SendText(`Somewhere in the realm, a <ansi fg="mobname">loot goblin</ansi> appears!`)
			slog.Info(`Loot Goblin Spawn`, `roundNumber`, util.GetRoundCount(), `forced`, true)
			gRoom.Prepare(false) // Make sure the loot goblin spawns.
		}
		return true, nil
	}

	mobId := mobs.MobIdByName(rest)

	if mobId < 1 {
		mobIdInt, _ := strconv.Atoi(rest)
		mobId = mobs.MobId(mobs.MobId(mobIdInt))
	}

	if mobId > 0 {
		if mob := mobs.NewMobById(mobId, room.RoomId); mob != nil {
			room.AddMob(mob.InstanceId)

			user.SendText(
				fmt.Sprintf(`You wave your hands around and <ansi fg="mobname">%s</ansi> appears in the air and falls to the ground.`, mob.Character.Name),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> waves their hands around and <ansi fg="mobname">%s</ansi> appears in the air and falls to the ground.`, user.Character.Name, mob.Character.Name),
				user.UserId,
			)

			return true, nil
		}
	}

	user.SendText(
		fmt.Sprintf(`Mob <ansi fg="mobname">%s</ansi> not found.`, rest),
	)

	return true, nil
}

func mob_Create(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`mob`, rest)

	mobCreateAnswerName := ``
	mobCreateAnswerRace := 0
	mobCreateAnswerZone := ``
	mobCreateAnswerDescription := `Not much to look at, really.`
	mobCreateAnswerScriptTemplate := false

	if isNew {
		user.SendText(``)
		user.SendText(fmt.Sprintf(`Lets get a little info first.%s`, term.CRLFStr))
	}

	//
	// Name Selection
	//

	question := cmdPrompt.Ask(`What will the mob be called?`, []string{}, `_`)
	if !question.Done {
		return true, nil
	}

	if question.Response == `_` {
		user.SendText("Aborting...")
		user.ClearPrompt()
		return true, nil
	}

	mobCreateAnswerName = question.Response

	//
	// Race Selection
	//
	allRaces := races.GetRaces()

	raceOptions := []templates.NameDescription{}
	for _, r := range allRaces {
		raceOptions = append(raceOptions, templates.NameDescription{
			Name:        r.Name,
			Description: r.Description,
		})
	}

	sort.SliceStable(raceOptions, func(i, j int) bool {
		return raceOptions[i].Name < raceOptions[j].Name
	})

	question = cmdPrompt.Ask(`What race will the mob be?`, []string{}, `_`)
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list", raceOptions)
		user.SendText(tplTxt)
		user.SendText(`  <ansi fg="black-bold">Enter <ansi fg="command">help {racename}</ansi> or <ansi fg="command">help {number}</ansi> for details.</ansi>`)
		user.SendText(``)
		return true, nil
	}

	if question.Response == `_` {
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

	for _, r := range allRaces {
		if strings.EqualFold(r.Name, raceNameSelection) {
			mobCreateAnswerRace = r.RaceId
		}
	}

	if mobCreateAnswerRace == 0 {
		question.RejectResponse()

		tplTxt, _ := templates.Process("tables/numbered-list", raceOptions)
		user.SendText(tplTxt)
		user.SendText(`  <ansi fg="black-bold">Enter <ansi fg="command">help {racename}</ansi> or <ansi fg="command">help {number}</ansi> for details.</ansi>`)
		user.SendText(``)

		return true, nil
	}

	//
	// Zone Selection
	//
	allZones := rooms.GetAllZoneNames()

	zoneOptions := []templates.NameDescription{}
	for _, z := range allZones {
		zoneOptions = append(zoneOptions, templates.NameDescription{
			Name:        z,
			Description: ``,
		})
	}

	sort.SliceStable(zoneOptions, func(i, j int) bool {
		return zoneOptions[i].Name < zoneOptions[j].Name
	})

	question = cmdPrompt.Ask(`What zone is this mob from?`, []string{}, `_`)
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list", zoneOptions)
		user.SendText(tplTxt)
		return true, nil
	}

	if question.Response == `_` {
		user.SendText("Aborting...")
		user.ClearPrompt()
		return true, nil
	}

	zoneNameSelection := question.Response
	if restNum, err := strconv.Atoi(zoneNameSelection); err == nil {
		if restNum > 0 && restNum <= len(zoneOptions) {
			zoneNameSelection = allZones[restNum-1]
		}
	}

	for _, z := range allZones {
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
	question = cmdPrompt.Ask(`Enter a description for the mob:`, []string{}, `_`)
	if !question.Done {
		return true, nil
	}

	if question.Response != `_` {
		mobCreateAnswerDescription = question.Response
	}

	//
	// Quest Script?
	//
	question = cmdPrompt.Ask(`Create with a default quest script?`, []string{`y`, `n`}, `n`)
	if !question.Done {
		return true, nil
	}

	mobCreateAnswerScriptTemplate = question.Response == `y`

	//
	// Confirm?
	//
	question = cmdPrompt.Ask(`Does this look correct?`, []string{`y`, `n`}, `n`)
	if !question.Done {

		user.SendText(`  <ansi fg="yellow-bold">Name:</ansi>    <ansi fg="white-bold">` + mobCreateAnswerName + `</ansi>`)
		user.SendText(`  <ansi fg="yellow-bold">Race:</ansi>    <ansi fg="white-bold">` + strconv.Itoa(mobCreateAnswerRace) + ` (` + raceNameSelection + `)</ansi>`)
		user.SendText(`  <ansi fg="yellow-bold">Zone:</ansi>    <ansi fg="white-bold">` + mobCreateAnswerZone + `</ansi>`)
		user.SendText(`  <ansi fg="yellow-bold">Desc:</ansi>    <ansi fg="white-bold">` + mobCreateAnswerDescription + `</ansi>`)
		user.SendText(`  <ansi fg="yellow-bold">Script:</ansi>  <ansi fg="white-bold">` + strconv.FormatBool(mobCreateAnswerScriptTemplate) + `</ansi>`)

		return true, nil
	}

	user.ClearPrompt()

	if question.Response != `y` {
		user.SendText("Aborting...")
		return true, nil
	}

	mobId, err := mobs.CreateNewMobFile(mobCreateAnswerName, mobCreateAnswerRace, mobCreateAnswerZone, mobCreateAnswerDescription, mobCreateAnswerScriptTemplate)

	if err != nil {
		user.SendText("Error: " + err.Error())
		return true, nil
	}

	mobInst := mobs.GetMobSpec(mobId)

	user.SendText(``)
	user.SendText(`  <ansi bg="red" fg="white-bold">MOB CREATED</ansi>`)
	user.SendText(``)
	user.SendText(`  <ansi fg="yellow-bold">File Path:</ansi>   <ansi fg="white-bold">` + mobInst.Filepath() + `</ansi>`)
	if mobCreateAnswerScriptTemplate {
		user.SendText(`  <ansi fg="yellow-bold">Script Path:</ansi> <ansi fg="white-bold">` + mobInst.GetScriptPath() + `</ansi>`)
	}
	user.SendText(``)
	user.SendText(`  <ansi fg="black-bold">note: Try <ansi fg="command">mob spawn ` + mobCreateAnswerName + `</ansi> to test it.`)

	return true, nil
}
