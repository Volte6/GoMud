package usercommands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/mutators"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Zone(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	handled := true

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.zone", nil)
		user.SendText(infoOutput)

		return handled, nil
	}

	zoneCmd := strings.ToLower(args[0])
	args = args[1:]

	// Interactive Editing
	if zoneCmd == `edit` {
		return zone_Edit(``, user, room, flags)
	}

	zoneConfig := rooms.GetZoneConfig(room.Zone)
	if zoneConfig == nil {
		user.SendText(fmt.Sprintf(`Couldn't find zone info for <ansi fg="red">%s</ansi>`, room.Zone))
		return true, nil
	}

	if zoneCmd == `info` {

		user.SendText(``)
		user.SendText(fmt.Sprintf(`<ansi fg="yellow-bold">Zone Config for <ansi fg="red">%s</ansi></ansi>`, room.Zone))
		user.SendText(fmt.Sprintf(`  <ansi fg="yellow-bold">Root Room Id:    </ansi> <ansi fg="red">%d</ansi>`, zoneConfig.RoomId))

		if zoneConfig.MobAutoScale.Maximum == 0 {
			user.SendText(`  <ansi fg="yellow-bold">Mob AutoScale:</ansi>    <ansi fg="red">[disabled]</ansi>`)
		} else {
			user.SendText(fmt.Sprintf(`  <ansi fg="yellow-bold">Mob AutoScale:</ansi>    <ansi fg="red">%d</ansi> - <ansi fg="red">%d</ansi>`, zoneConfig.MobAutoScale.Minimum, zoneConfig.MobAutoScale.Maximum))
		}

		user.SendText(``)

		return true, nil
	}

	// Everthing after this point requires additional args
	if len(args) < 1 {
		user.SendText(`Not enough arguments provided.`)
		return true, nil
	}

	if zoneCmd == `set` {

		setWhat := args[0]

		args = args[1:]

		if setWhat == `autoscale` {
			if len(args) < 2 {
				user.SendText(`Use <ansi fg="command">zone set autoscale 0 0</ansi> to clear autoscaling.`)
				return true, nil
			}

			min, _ := strconv.Atoi(args[0])
			max, _ := strconv.Atoi(args[1])

			if min < 0 || max < 0 {
				user.SendText(`Min/Max can't be less than zero.`)
				return true, nil
			}

			zoneConfig.MobAutoScale.Minimum = min
			zoneConfig.MobAutoScale.Maximum = max
			zoneConfig.Validate()

			user.SendText(`Done!`)
			return true, nil
		}

	}

	return true, nil
}

func zone_Edit(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	originalZoneConfig := rooms.GetZoneConfig(room.Zone)
	if originalZoneConfig == nil {
		user.SendText(`Could not find zone config.`)
		return true, nil
	}

	// Make a copy that we'll edit
	editZoneConfig := *originalZoneConfig

	allZoneMutators := []string{}
	for _, roomMut := range editZoneConfig.Mutators {
		allZoneMutators = append(allZoneMutators, roomMut.MutatorId)
	}

	cmdPrompt, _ := user.StartPrompt(`zone edit`, rest)

	selectedMutatorList := []string{}
	if muts, ok := cmdPrompt.Recall(`mutators`); ok {
		selectedMutatorList = muts.([]string)
	} else {
		if len(selectedMutatorList) == 0 {
			selectedMutatorList = append(selectedMutatorList, allZoneMutators...)
		}
	}

	// Keep track of the state
	cmdPrompt.Store(`mutators`, selectedMutatorList)

	selectedMutatorLookup := map[string]bool{}
	for _, mutId := range selectedMutatorList {
		selectedMutatorLookup[mutId] = true
	}

	mutatorOptions := []templates.NameDescription{}

	for _, mutId := range mutators.GetAllMutatorIds() {
		marked := false
		if _, ok := selectedMutatorLookup[mutId]; ok {
			marked = true
		}

		mutatorOptions = append(mutatorOptions, templates.NameDescription{Id: mutId, Marked: marked, Name: mutId})

	}

	sort.SliceStable(mutatorOptions, func(i, j int) bool {
		return mutatorOptions[i].Name < mutatorOptions[j].Name
	})

	question := cmdPrompt.Ask(`Select a mutator to add/remove, or nothing to continue:`, []string{}, `0`)
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list-doubled", mutatorOptions)
		user.SendText(tplTxt)
		return true, nil
	}

	if question.Response != `0` {

		mutatorSelected := ``

		if restNum, err := strconv.Atoi(question.Response); err == nil {
			if restNum > 0 && restNum <= len(mutatorOptions) {
				mutatorSelected = mutatorOptions[restNum-1].Id.(string)
			}
		}

		if mutatorSelected == `` {
			for _, b := range mutatorOptions {
				if strings.EqualFold(b.Name, question.Response) {
					mutatorSelected = b.Id.(string)
					break
				}
			}
		}

		if mutatorSelected == `` {

			user.SendText("Invalid selection.")
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-list-doubled", mutatorOptions)
			user.SendText(tplTxt)
			return true, nil
		}

		if _, ok := selectedMutatorLookup[mutatorSelected]; ok {

			delete(selectedMutatorLookup, mutatorSelected)
			for idx, mutId := range selectedMutatorList {
				if mutId == mutatorSelected {
					selectedMutatorList = append(selectedMutatorList[0:idx], selectedMutatorList[idx+1:]...)
					break
				}
			}

		} else {

			selectedMutatorList = append(selectedMutatorList, mutatorSelected)
			selectedMutatorLookup[mutatorSelected] = true

		}

		cmdPrompt.Store(`mutators`, selectedMutatorList)

		question.RejectResponse()

		for idx, data := range mutatorOptions {
			_, data.Marked = selectedMutatorLookup[data.Id.(string)]
			mutatorOptions[idx] = data
		}

		tplTxt, _ := templates.Process("tables/numbered-list-doubled", mutatorOptions)
		user.SendText(tplTxt)
		return true, nil

	}

	//
	// Auto-scaling Options
	//
	{

		question := cmdPrompt.Ask(`Mob Autoscaling MINIMUM level?`, []string{strconv.Itoa(editZoneConfig.MobAutoScale.Minimum)}, strconv.Itoa(editZoneConfig.MobAutoScale.Minimum))
		if !question.Done {
			return true, nil
		}
		editZoneConfig.MobAutoScale.Minimum, _ = strconv.Atoi(question.Response)

		question = cmdPrompt.Ask(`Mob Autoscaling MAXIMUM level?`, []string{strconv.Itoa(editZoneConfig.MobAutoScale.Maximum)}, strconv.Itoa(editZoneConfig.MobAutoScale.Maximum))
		if !question.Done {
			return true, nil
		}
		editZoneConfig.MobAutoScale.Maximum, _ = strconv.Atoi(question.Response)

	}

	//
	// Music Options
	//
	{

		question := cmdPrompt.Ask(`Should the zone have music?`, []string{`yes`, `no`}, util.BoolYN(editZoneConfig.MusicFile != ``))
		if !question.Done {
			return true, nil
		}

		if question.Response == `yes` {

			question := cmdPrompt.Ask(`Zone music file?`, []string{editZoneConfig.MusicFile}, editZoneConfig.MusicFile)
			if !question.Done {
				return true, nil
			}
			editZoneConfig.MusicFile = question.Response

		} else {
			editZoneConfig.MusicFile = ``
		}

	}

	//
	// Done editing. Save results
	//
	editZoneConfig.Mutators = mutators.MutatorList{}
	for _, mutId := range selectedMutatorList {
		editZoneConfig.Mutators = append(editZoneConfig.Mutators, mutators.Mutator{MutatorId: mutId})
	}

	// Make sure the edited zone config's roomId gets the changes.
	if r := rooms.LoadRoom(editZoneConfig.RoomId); r != nil {
		r.ZoneConfig = editZoneConfig
	}

	// If the root zone room has been changed, clear the original rooms zone config.
	if originalZoneConfig.RoomId != editZoneConfig.RoomId {
		if r := rooms.LoadRoom(originalZoneConfig.RoomId); r != nil {
			room.ZoneConfig = rooms.ZoneConfig{}
		}
	}

	user.SendText(``)
	user.SendText(`Changes saved.`)
	user.SendText(``)

	user.ClearPrompt()

	return true, nil
}
