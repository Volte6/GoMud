package usercommands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/exit"
	"github.com/volte6/gomud/internal/gamelock"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mutators"
	"github.com/volte6/gomud/internal/parties"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
)

func Room(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	handled := true

	// args should look like one of the following:
	// info <optional room id>
	// <move to room id>
	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.room", nil)
		user.SendText(infoOutput)

		return handled, nil
	}

	var roomId int = 0
	roomCmd := strings.ToLower(args[0])

	// Interactive Editing
	if roomCmd == `edit` {

		if rest == `edit container` || rest == `edit containers` {
			return room_Edit_Containers(``, user, room, flags)
		}

		if rest == `edit exit` || rest == `edit exits` {
			return room_Edit_Exits(``, user, room, flags)
		}

		if rest == `edit mutator` || rest == `edit mutators` {
			return room_Edit_Mutators(``, user, room, flags)
		}

		user.SendText(`<ansi fg="red">edit WHAT?</ansi> Try:`)
		user.SendText(`    <ansi fg="command">room edit containers</ansi>`)
		user.SendText(`    <ansi fg="command">room edit exits</ansi>`)
		user.SendText(`    <ansi fg="command">room edit mutators</ansi>`)

		return true, nil
	}

	if roomCmd == `noun` || roomCmd == `nouns` {

		// room noun chair "a chair for sitting"
		if len(args) > 2 {
			noun := args[1]
			description := strings.Join(args[2:], ` `)

			if room.Nouns == nil {
				room.Nouns = map[string]string{}
			}
			room.Nouns[noun] = description

			user.SendText(`Noun Added:`)
			user.SendText(fmt.Sprintf(`<ansi fg="noun">%s</ansi> - %s`, strings.Repeat(` `, 20-len(noun))+noun, description))

			return true, nil
		}

		// room noun chair
		if len(args) == 2 || (len(args) == 3 && len(args[2]) == 0) {

			if _, ok := room.Nouns[args[1]]; ok {
				delete(room.Nouns, args[1])
				user.SendText(`Noun deleted.`)
			} else {
				user.SendText(`Noun not found.`)
			}

			return true, nil
		}

		// room noun
		// room nouns
		user.SendText(`Room Nouns:`)
		for noun, description := range room.Nouns {
			user.SendText(fmt.Sprintf(`<ansi fg="noun">%s</ansi> - %s`, strings.Repeat(` `, 20-len(noun))+noun, description))
		}
		return true, nil
	}

	if roomCmd == "copy" && len(args) >= 3 {

		property := args[1]

		if property == "spawninfo" {
			sourceRoom, _ := strconv.Atoi(args[2])
			// copy something from another room
			if sourceRoom := rooms.LoadRoom(sourceRoom); sourceRoom != nil {

				room.SpawnInfo = sourceRoom.SpawnInfo
				rooms.SaveRoom(*room)

				user.SendText("Spawn info copied/overwritten.")
			}
		}

		if property == "idlemessages" {
			sourceRoom, _ := strconv.Atoi(args[2])
			// copy something from another room
			if sourceRoom := rooms.LoadRoom(sourceRoom); sourceRoom != nil {

				room.IdleMessages = append(room.IdleMessages, sourceRoom.IdleMessages...)
				rooms.SaveRoom(*room)

				user.SendText("IdleMessages copied/overwritten.")
			}
		}

		if property == "mutator" || property == "mutators" {
			sourceRoom, _ := strconv.Atoi(args[2])
			// copy something from another room
			if sourceRoom := rooms.LoadRoom(sourceRoom); sourceRoom != nil {

				room.Mutators = append(room.Mutators, sourceRoom.Mutators...)
				rooms.SaveRoom(*room)

				user.SendText("Mutators copied/overwritten.")
			}
		}

	} else if roomCmd == "info" {
		if len(args) == 1 {
			roomId = room.RoomId
		} else {
			roomId, _ = strconv.Atoi(args[1])
		}

		targetRoom := rooms.LoadRoom(roomId)
		if targetRoom == nil {
			user.SendText(fmt.Sprintf("Room %d not found.", roomId))
			return false, fmt.Errorf("room %d not found", roomId)
		}

		roomInfo := map[string]any{
			`room`: targetRoom,
			`zone`: rooms.GetZoneConfig(targetRoom.Zone),
		}

		infoOutput, _ := templates.Process("admincommands/ingame/roominfo", roomInfo)
		user.SendText(infoOutput)

	} else if len(args) >= 2 && roomCmd == "exit" {

		// exit west 159 <- Create/change exit with roomId as target room
		// exit up climb <- Rename exit

		direction := strings.ToLower(args[1])
		roomId = 0
		var numError error = nil
		exitRename := ``

		if len(args) > 2 {
			roomId, numError = strconv.Atoi(args[2])
			if numError != nil {
				exitRename = args[2]
			}
		}

		// Will be erasing it.
		if len(args) < 3 { // If NO room number/name supplied, delete
			if _, ok := room.Exits[direction]; !ok {
				user.SendText(fmt.Sprintf("Exit %s does not exist.", direction))
				return handled, nil
			}
			delete(room.Exits, direction)
			return handled, nil
		}

		if currentExit, ok := room.Exits[direction]; ok {
			user.SendText(fmt.Sprintf("Exit %s already exists (overwriting).", direction))

			if exitRename != `` {
				delete(room.Exits, direction)
				room.Exits[exitRename] = currentExit

				user.SendText(fmt.Sprintf("Exit %s renamed to %s.", direction, exitRename))
				return true, nil
			}
		}

		targetRoom := rooms.LoadRoom(roomId)
		if targetRoom == nil {
			err := fmt.Errorf(`room %d not found`, roomId)
			user.SendText(err.Error())
			return handled, nil
		}

		rooms.ConnectRoom(room.RoomId, targetRoom.RoomId, direction)
		user.SendText(fmt.Sprintf("Exit %s added.", direction))

	} else if len(args) >= 2 && roomCmd == "secretexit" {

		direction := args[1]
		if exit, ok := room.Exits[direction]; ok {
			if exit.Secret {
				exit.Secret = false
				room.Exits[direction] = exit
				rooms.SaveRoom(*room)
				user.SendText(fmt.Sprintf("Exit %s secrecy REMOVED.", direction))
			} else {
				exit.Secret = true
				room.Exits[direction] = exit
				rooms.SaveRoom(*room)
				user.SendText(fmt.Sprintf("Exit %s secrecy ADDED.", direction))
			}
		} else {
			user.SendText(fmt.Sprintf("Exit %s not found.", direction))
		}

	} else if len(args) >= 2 && roomCmd == "set" {

		propertyName := args[1]
		propertyValue := ``
		if len(args) > 2 {
			propertyValue = strings.Join(args[2:], ` `)
		}

		propertyValue = strings.Trim(propertyValue, `"`)

		if propertyName == "mutator" || propertyName == "mutators" {

			if propertyValue == `` { // If none specified, list all mutators

				user.SendText(`<ansi fg="table-title">Mutators:</ansi>`)
				if len(room.Mutators) == 0 {
					user.SendText(`  None.`)
				}
				for _, mut := range room.Mutators {
					user.SendText(`  <ansi fg="mutator">` + mut.MutatorId + `</ansi>`)
				}
				user.SendText(``)

			} else { // Otherwise, toggle the mentioned mutator on/off

				user.SendText(``)

				if !mutators.IsMutator(propertyValue) {
					user.SendText(`<ansi fg="table-title"><ansi fg="mutator">` + propertyValue + `</ansi> is an invalid mutator id.</ansi>`)
					user.SendText(`<ansi fg="table-title">  Here is a list of valid mutator id's:</ansi>`)
					for _, name := range mutators.GetAllMutatorIds() {
						user.SendText(`    <ansi fg="mutator">` + name + `</ansi>`)
					}
				} else if room.Mutators.Remove(propertyValue) {
					user.SendText(`<ansi fg="table-title">Mutator <ansi fg="mutator">` + propertyValue + `</ansi> Removed.</ansi>`)
				} else if room.Mutators.Add(propertyValue) {
					user.SendText(`<ansi fg="table-title">Mutator <ansi fg="mutator">` + propertyValue + `</ansi> Added.</ansi>`)
				}

				user.SendText(``)
			}

			return true, nil
		}

		if propertyName == "spawninfo" {
			if propertyValue == `clear` {
				room.SpawnInfo = room.SpawnInfo[:0]
				rooms.SaveRoom(*room)
			}

		} else if propertyName == "title" {
			if propertyValue == `` {
				propertyValue = `[no title]`
			}
			room.Title = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "description" {
			if propertyValue == `` {
				propertyValue = `[no description]`
			}
			propertyValue = strings.ReplaceAll(propertyValue, `\n`, "\n")
			room.Description = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "idlemessages" {
			room.IdleMessages = []string{}
			for _, idleMsg := range strings.Split(propertyValue, ";") {
				idleMsg = strings.TrimSpace(idleMsg)
				if len(idleMsg) < 1 {
					continue
				}
				room.IdleMessages = append(room.IdleMessages, idleMsg)
			}
			rooms.SaveRoom(*room)
		} else if propertyName == "symbol" || propertyName == "mapsymbol" {
			room.MapSymbol = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "legend" || propertyName == "maplegend" {
			room.MapLegend = propertyValue
			rooms.SaveRoom(*room)
		} else if propertyName == "zone" {
			// Try moving it to the new zone.
			if err := rooms.MoveToZone(room.RoomId, propertyValue); err != nil {
				user.SendText(err.Error())
				return handled, nil
			}

		} else if propertyName == "biome" {
			room.Biome = strings.ToLower(propertyValue)
		} else {
			user.SendText(
				`Invalid property provided to <ansi fg="command">room set</ansi>.`,
			)
			return false, fmt.Errorf("room %d not found", roomId)
		}

	} else {

		var gotoRoomId int = 0
		var numError error = nil

		if deltaD, ok := rooms.DirectionDeltas[roomCmd]; ok {

			rGraph := rooms.NewRoomGraph(100, 100, 0, rooms.MapModeAll)
			err := rGraph.Build(user.Character.RoomId, nil)
			if err != nil {
				user.SendText(err.Error())
				return true, nil
			}

			map2D, cX, cY := rGraph.Generate2DMap(61, 61, user.Character.RoomId)
			if len(map2D) < 1 {
				user.SendText("Error generating a 2d map")
				return true, nil
			}

			for i := 1; i <= 30; i++ {
				dy := deltaD.Dy * i
				dx := deltaD.Dx * i
				if cY+dy < len(map2D) && cX+dx < len(map2D[0]) {
					if cY+dy >= 0 && cX+dx >= 0 {
						if map2D[cY+dy][cX+dx] != nil {
							gotoRoomId = map2D[cY+dy][cX+dx].RoomId
							break
						}
					}
				}
			}

			//dirDelta.Dx
			//dirDelta.Dy
		} else {
			// move to a new room
			gotoRoomId, numError = strconv.Atoi(args[0])
		}

		if numError == nil {

			previousRoomId := user.Character.RoomId

			if err := rooms.MoveToRoom(user.UserId, gotoRoomId); err != nil {
				user.SendText(err.Error())

			} else {

				scripting.TryRoomScriptEvent(`onExit`, user.UserId, previousRoomId)

				user.SendText(fmt.Sprintf("Moved to room %d.", gotoRoomId))

				gotoRoom := rooms.LoadRoom(gotoRoomId)
				gotoRoom.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> appears in a flash of light!`, user.Character.Name),
					user.UserId,
				)

				if party := parties.Get(user.UserId); party != nil {

					newRoom := rooms.LoadRoom(gotoRoomId)
					for _, uid := range room.GetPlayers() {
						if party.IsMember(uid) {

							partyUser := users.GetByUserId(uid)
							if partyUser == nil {
								continue
							}

							rooms.MoveToRoom(partyUser.UserId, gotoRoomId)
							user.SendText(fmt.Sprintf("Moved to room %d.", gotoRoomId))
							room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> appears in a flash of light!`, partyUser.Character.Name), partyUser.UserId)

							for _, mInstanceId := range room.GetMobs(rooms.FindCharmed) {
								if mob := mobs.GetInstance(mInstanceId); mob != nil {
									if mob.Character.IsCharmed(partyUser.UserId) {
										room.RemoveMob(mob.InstanceId)
										newRoom.AddMob(mob.InstanceId)
									}
								}
							}
						}
					}
				}

				Look(``, user, gotoRoom, flags)

				scripting.TryRoomScriptEvent(`onEnter`, user.UserId, gotoRoomId)

			}
		} else {
			user.SendText(fmt.Sprintf("Invalid room command: %s", args[0]))
		}
	}

	return handled, nil
}

func room_Edit_Containers(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// This basic struct will be used to keep track of what we're editing
	type ContainerEdit struct {
		Name      string
		NameNew   string
		Container rooms.Container
		Exists    bool
	}

	containerOptions := []templates.NameDescription{}

	for name, c := range room.Containers {

		// If it's ephemeral, don't bother.
		if c.DespawnRound != 0 {
			continue
		}

		containerOption := templates.NameDescription{Name: name}

		if c.Lock.Difficulty > 0 {
			containerOption.Description += fmt.Sprintf(`[Lvl %d Lock] `, c.Lock.Difficulty)
		}

		if len(c.Recipes) > 0 {
			containerOption.Description += fmt.Sprintf(`[%d Recipe(s)] `, len(c.Recipes))
		}

		containerOptions = append(containerOptions, containerOption)

	}

	// Must sort since maps will often change between iterations
	sort.SliceStable(containerOptions, func(i, j int) bool {
		return containerOptions[i].Name < containerOptions[j].Name
	})

	//
	// Create a holder for container editing data
	//
	currentlyEditing := ContainerEdit{}

	cmdPrompt, _ := user.StartPrompt(`room edit containers`, rest)

	question := cmdPrompt.Ask(`Choose one:`, []string{`new`}, `new`)
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list", containerOptions)
		user.SendText(tplTxt)
		return true, nil
	}

	currentlyEditing.Name = question.Response

	if restNum, err := strconv.Atoi(currentlyEditing.Name); err == nil {
		if restNum > 0 && restNum <= len(containerOptions) {
			currentlyEditing.Name = containerOptions[restNum-1].Name
		}
	}

	for _, o := range containerOptions {
		if strings.EqualFold(o.Name, currentlyEditing.Name) {
			currentlyEditing.Name = o.Name
			break
		}
	}

	// Load the (possible) existing container
	currentlyEditing.Container, currentlyEditing.Exists = room.Containers[currentlyEditing.Name]

	// If they entered a container name...
	if currentlyEditing.Name != `new` {

		// Does the container name they entered not exist? Failure!
		if !currentlyEditing.Exists {
			user.SendText("Invalid option selected.")
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		// Since they picked a container that exists, lets get the question of delete out of the way immediately.
		question := cmdPrompt.Ask(`Delete this container?`, []string{`yes`, `no`}, `no`)
		if !question.Done {
			return true, nil
		}

		// Delete the container if that's what they want!
		if question.Response == `yes` {

			delete(room.Containers, currentlyEditing.Name)
			rooms.SaveRoom(*room)

			user.SendText(``)
			user.SendText(fmt.Sprintf(`<ansi fg="container">%s</ansi> deleted from the room.`, currentlyEditing.Name))
			user.SendText(``)

			user.ClearPrompt()
			return true, nil
		}

	}

	//
	// Name Selection
	//
	{
		// If they are creating a new container, we don't want that to become a viable container name, lets empty it
		if currentlyEditing.Name == `new` {
			currentlyEditing.Name = ``
		}

		// allow them to name/rename the container.
		question := cmdPrompt.Ask(`Choose a name for this container:`, []string{currentlyEditing.Name}, currentlyEditing.Name)
		if !question.Done {
			return true, nil
		}
		currentlyEditing.NameNew = question.Response

		// Make sure they aren't using any reserved names.
		if currentlyEditing.NameNew == `quit` || currentlyEditing.NameNew == `new` {
			user.SendText("Invalid new name selected.")
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		// Make sure the new name isn't a duplicate
		if currentlyEditing.Name != currentlyEditing.NameNew {
			if _, ok := room.Containers[currentlyEditing.NameNew]; ok {

				user.SendText(`<ansi fg="red">A container with that name already exists!</ansi>`)
				question.RejectResponse()
				return true, nil

			}
		}

	}

	//
	// Lock Options
	//
	{
		question := cmdPrompt.Ask(`Will this container be locked?`, []string{`yes`, `no`}, util.BoolYN(currentlyEditing.Container.Lock.Difficulty > 0))
		if !question.Done {
			return true, nil
		}

		if question.Response == `yes` {

			defaultDifficultyAnswer := ``
			if currentlyEditing.Container.Lock.Difficulty > 0 {
				defaultDifficultyAnswer = strconv.Itoa(int(currentlyEditing.Container.Lock.Difficulty))
			}

			question := cmdPrompt.Ask(`What difficulty will the lock be (2-32)?`, []string{defaultDifficultyAnswer}, defaultDifficultyAnswer)
			if !question.Done {
				return true, nil
			}

			difficultyInt, _ := strconv.Atoi(question.Response)

			// Make sure the provided difficulty is within acceptable range.
			if difficultyInt < 2 || difficultyInt > 32 {
				user.SendText("Difficulty must between 2 and 32, inclusive.")
				question.RejectResponse()
				return true, nil
			}

			currentlyEditing.Container.Lock.Difficulty = uint8(difficultyInt)

		} else {
			// reset the lock state if there is no lock.
			currentlyEditing.Container.Lock = gamelock.Lock{}
		}

		if currentlyEditing.Container.Lock.Difficulty > 0 {
			//
			// Lock Trap Options
			//
			question = cmdPrompt.Ask(`Will this lock have a trap?`, []string{`yes`, `no`}, util.BoolYN(len(currentlyEditing.Container.Lock.TrapBuffIds) > 0))
			if !question.Done {
				return true, nil
			}

			if question.Response == `yes` {

				selectedBuffList := []int{}
				if cb, ok := cmdPrompt.Recall(`trapBuffs`); ok {
					selectedBuffList = cb.([]int)
				}

				if len(selectedBuffList) == 0 {
					selectedBuffList = append(selectedBuffList, currentlyEditing.Container.Lock.TrapBuffIds...)
				}

				// Keep track of the state
				cmdPrompt.Store(`trapBuffs`, selectedBuffList)

				selectedBuffLookup := map[int]bool{}
				for _, bId := range selectedBuffList {
					selectedBuffLookup[bId] = true
				}

				buffOptions := []templates.NameDescription{}

				for _, buffId := range buffs.GetAllBuffIds() {
					if b := buffs.GetBuffSpec(buffId); b != nil {

						if b.Name == `empty` {
							continue
						}

						marked := false
						if _, ok := selectedBuffLookup[buffId]; ok {
							marked = true
						}

						buffOptions = append(buffOptions, templates.NameDescription{Id: buffId, Marked: marked, Name: b.Name})
					}
				}

				sort.SliceStable(buffOptions, func(i, j int) bool {
					return buffOptions[i].Name < buffOptions[j].Name
				})

				question := cmdPrompt.Ask(`Select a buff to add to the trap, or nothing to continue:`, []string{}, `0`)
				if !question.Done {
					tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions)
					user.SendText(tplTxt)
					return true, nil
				}

				buffSelected := question.Response

				if buffSelected != `0` {

					buffSelectedInt := 0

					if restNum, err := strconv.Atoi(buffSelected); err == nil {
						if restNum > 0 && restNum <= len(buffOptions) {
							buffSelectedInt = buffOptions[restNum-1].Id.(int)
						}
					}

					if buffSelectedInt == 0 {
						for _, b := range buffOptions {
							if strings.EqualFold(b.Name, buffSelected) {
								buffSelectedInt = b.Id.(int)
								break
							}
						}
					}

					if buffSelectedInt == 0 {

						user.SendText("Invalid selection.")
						question.RejectResponse()

						tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions)
						user.SendText(tplTxt)
						return true, nil
					}

					if _, ok := selectedBuffLookup[buffSelectedInt]; ok {

						delete(selectedBuffLookup, buffSelectedInt)
						for idx, buffId := range selectedBuffList {
							if buffId == buffSelectedInt {
								selectedBuffList = append(selectedBuffList[0:idx], selectedBuffList[idx+1:]...)
								break
							}
						}

					} else {

						selectedBuffList = append(selectedBuffList, buffSelectedInt)
						selectedBuffLookup[buffSelectedInt] = true

					}

					cmdPrompt.Store(`trapBuffs`, selectedBuffList)

					question.RejectResponse()

					for idx, data := range buffOptions {
						_, data.Marked = selectedBuffLookup[data.Id.(int)]
						buffOptions[idx] = data
					}

					tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions)
					user.SendText(tplTxt)
					return true, nil

				}

			}

			if cb, ok := cmdPrompt.Recall(`trapBuffs`); ok {
				currentlyEditing.Container.Lock.TrapBuffIds = cb.([]int)
			}

			if currentlyEditing.Container.Lock.RelockInterval == `` {
				currentlyEditing.Container.Lock.RelockInterval = gamelock.DefaultRelockTime
			}

			question = cmdPrompt.Ask(`How long until it automatically relocks?`, []string{currentlyEditing.Container.Lock.RelockInterval}, currentlyEditing.Container.Lock.RelockInterval)
			if !question.Done {
				return true, nil
			}

			currentlyEditing.Container.Lock.RelockInterval = question.Response

			// If the default time is chosen, can just leave it blank.
			if currentlyEditing.Container.Lock.RelockInterval == gamelock.DefaultRelockTime {
				currentlyEditing.Container.Lock.RelockInterval = ``
			}

		}
	}

	//
	// Recipe Options
	//
	{
		question := cmdPrompt.Ask(`Will this container have recipes?`, []string{`yes`, `no`}, util.BoolYN(len(currentlyEditing.Container.Recipes) > 0))
		if !question.Done {
			return true, nil
		}

		if question.Response == `yes` {

			currentRecipes := map[int][]int{}
			if cr, ok := cmdPrompt.Recall(`recipes`); ok {
				currentRecipes = cr.(map[int][]int)
			}

			if len(currentRecipes) == 0 {
				for k, v := range currentlyEditing.Container.Recipes {
					currentRecipes[k] = append([]int{}, v...)
				}
			}

			recipeNow := 0
			if rNow, ok := cmdPrompt.Recall(`recipeNow`); ok {
				recipeNow = rNow.(int)
			}

			if recipeNow != 0 && items.GetItemSpec(recipeNow) == nil {
				user.SendText(`<ansi fg="red">Invalid selection.</ansi>`)
				question.RejectResponse()
				return true, nil
			}

			// Keep track of the state
			cmdPrompt.Store(`recipes`, currentRecipes)
			cmdPrompt.Store(`recipeNow`, recipeNow)

			// Select recipe to modify
			if _, ok := currentRecipes[recipeNow]; !ok {
				recipeOptions := []templates.NameDescription{}
				for productItemId, recipeItemList := range currentRecipes {

					itm := items.New(productItemId)
					productName := fmt.Sprintf(`%d (%s)`, productItemId, itm.DisplayName())

					allRequiredItems := []string{}
					for _, iId := range recipeItemList {
						itm := items.New(iId)
						allRequiredItems = append(allRequiredItems, fmt.Sprintf(`%d (%s)`, iId, itm.DisplayName()))
					}

					recipeOptions = append(recipeOptions,
						templates.NameDescription{
							Id:          productItemId,
							Marked:      recipeNow == productItemId,
							Name:        productName,
							Description: strings.Join(allRequiredItems, `, `),
						})

				}

				recipeOptions = append(recipeOptions,
					templates.NameDescription{
						Id:          0,
						Marked:      false,
						Name:        `new`,
						Description: `create a new recipe`,
					})

				recipeOptions = append(recipeOptions,
					templates.NameDescription{
						Id:          -1,
						Marked:      false,
						Name:        `skip`,
						Description: `skip this step`,
					})

				question := cmdPrompt.Ask(`Modify which (or new)?`, []string{`skip`}, `skip`)
				if !question.Done {
					tplTxt, _ := templates.Process("tables/numbered-list", recipeOptions)
					user.SendText(tplTxt)
					return true, nil
				}

				recipeSelected := question.Response
				if restNum, err := strconv.Atoi(recipeSelected); err == nil {
					if restNum > 0 && restNum <= len(recipeOptions) {
						recipeNow = recipeOptions[restNum-1].Id.(int)
					}
				}

				if recipeNow == 0 {
					for _, b := range recipeOptions {
						if strings.EqualFold(b.Name, recipeSelected) {
							recipeNow = b.Id.(int)
							break
						}
					}
				}

				if question.Response == `new` {

					question := cmdPrompt.Ask(`What itemId will be created?`, []string{})
					if !question.Done {
						return true, nil
					}

					itemIdInt, _ := strconv.Atoi(question.Response)
					if items.GetItemSpec(itemIdInt) == nil {

						user.SendText("Invalid itemId.")
						question.RejectResponse()

						return true, nil
					}

					if _, ok := currentRecipes[itemIdInt]; !ok {
						currentRecipes[itemIdInt] = []int{}
					}

					recipeNow = itemIdInt

					// Keep track of the state
					cmdPrompt.Store(`recipes`, currentRecipes)
					cmdPrompt.Store(`recipeNow`, recipeNow)
				}
			}

			// If they're editing a recipe, lets add ingredients
			if recipeNow != -1 {

				neededItems := map[int]int{}
				for _, inputItemId := range currentRecipes[recipeNow] {
					neededItems[inputItemId] = neededItems[inputItemId] + 1
				}

				question = cmdPrompt.Ask(`Enter an itemId to add to the recipe, or nothing to continue:`, []string{``}, `skip`)
				if !question.Done {
					// They have a recipe to modify, ask for item id's
					user.SendText(``)
					user.SendText(`<ansi fg="cyan">Positive numbers add items, negative numbers remove items.</ansi>`)

					room_Edit_Containers_SendRecipes(user, recipeNow, neededItems)

					return true, nil
				}

				if question.Response != `skip` {

					removeItem := false
					if question.Response[0] == '-' {
						removeItem = true
						question.Response = question.Response[1:]
					}

					recipeAdjustment := items.FindItem(question.Response)

					if itemSpec := items.GetItemSpec(recipeAdjustment); itemSpec == nil {
						user.SendText(`<ansi fg="red">Invalid ItemId provided.</ansi>`)

						room_Edit_Containers_SendRecipes(user, recipeNow, neededItems)

						question.RejectResponse()
						return true, nil
					}

					if removeItem {

						for idx, itemId := range currentRecipes[recipeNow] {

							if itemId == recipeAdjustment {
								currentRecipes[recipeNow] = append(currentRecipes[recipeNow][0:idx], currentRecipes[recipeNow][idx+1:]...)

								neededItems[recipeAdjustment] -= 1

								if neededItems[recipeAdjustment] == 0 {
									delete(neededItems, recipeAdjustment)
								}

								break
							}

						}

					} else {
						currentRecipes[recipeNow] = append(currentRecipes[recipeNow], recipeAdjustment)
						neededItems[recipeAdjustment] += 1
					}

					// Keep track of the state
					cmdPrompt.Store(`recipes`, currentRecipes)
					cmdPrompt.Store(`recipeNow`, recipeNow)

					room_Edit_Containers_SendRecipes(user, recipeNow, neededItems)

					question.RejectResponse()
					return true, nil

				}

			}

			if allRecipes, ok := cmdPrompt.Recall(`recipes`); ok {
				currentlyEditing.Container.Recipes = allRecipes.(map[int][]int)

				for i, itms := range currentlyEditing.Container.Recipes {
					if len(itms) == 0 {
						delete(currentlyEditing.Container.Recipes, i)
					}
				}
			}

		} else {
			clear(currentlyEditing.Container.Recipes)
		}

	}

	//
	// Done editing. Save results
	//
	if currentlyEditing.Name != `` {
		delete(room.Containers, currentlyEditing.Name)
	}

	room.Containers[currentlyEditing.NameNew] = currentlyEditing.Container
	rooms.SaveRoom(*room)

	user.SendText(``)

	if currentlyEditing.Container.Lock.Difficulty > 0 {
		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, currentlyEditing.NameNew)
		user.SendText(fmt.Sprintf(`<ansi fg="red">To Create Key -  LockId: <ansi fg="231" bg="5">%s</ansi></ansi>`, lockId))

		seqString := ``
		for _, dir := range util.GetLockSequence(lockId, int(currentlyEditing.Container.Lock.Difficulty), string(configs.GetServerConfig().Seed)) {
			seqString += string(dir) + " "
		}
		user.SendText(fmt.Sprintf(`<ansi fg="red">To pick lock - Sequence: <ansi fg="green">%s</ansi></ansi>`, seqString))
	}

	user.SendText(``)
	user.SendText(`Changes saved.`)
	user.SendText(``)

	user.ClearPrompt()

	return true, nil
}

func room_Edit_Containers_SendRecipes(user *users.UserRecord, recipeResultItemId int, recipeItems map[int]int) {

	itm := items.New(recipeResultItemId)

	user.SendText(``)
	user.SendText(fmt.Sprintf(`    Current Recipe for %d (<ansi fg="itemname">%s</ansi>):`, recipeResultItemId, itm.DisplayName()))

	itemsList := []string{}
	for itemId, qty := range recipeItems {
		itm := items.New(itemId)
		itemsList = append(itemsList, fmt.Sprintf(`        <ansi fg="red">[x%d]</ansi> %d (<ansi fg="itemname">%s</ansi>)`, qty, itemId, itm.DisplayName()))
	}

	// Must sort since maps will often change between iterations
	sort.SliceStable(itemsList, func(i, j int) bool {
		return itemsList[i] < itemsList[j]
	})

	for _, txt := range itemsList {
		user.SendText(txt)
	}

	user.SendText(``)
}

func room_Edit_Exits(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// This basic struct will be used to keep track of what we're editing
	type ExitEdit struct {
		Name    string
		NameNew string
		Exit    exit.RoomExit
		Exists  bool
	}

	exitOptions := []templates.NameDescription{}

	for name, c := range room.Exits {

		exitOpt := templates.NameDescription{Name: name}

		if c.Lock.Difficulty > 0 {
			exitOpt.Description += fmt.Sprintf(`[Lvl %d Lock] `, c.Lock.Difficulty)
		}

		if c.Secret {
			exitOpt.Description += `[hidden] `
		}

		exitOptions = append(exitOptions, exitOpt)

	}

	// Must sort since maps will often change between iterations
	sort.SliceStable(exitOptions, func(i, j int) bool {
		return exitOptions[i].Name < exitOptions[j].Name
	})

	//
	// Create a holder for exit editing data
	//
	currentlyEditing := ExitEdit{}

	cmdPrompt, _ := user.StartPrompt(`room edit exits`, rest)

	question := cmdPrompt.Ask(`Choose one:`, []string{`new`}, `new`)
	if !question.Done {
		tplTxt, _ := templates.Process("tables/numbered-list", exitOptions)
		user.SendText(tplTxt)
		return true, nil
	}

	currentlyEditing.Name = question.Response

	if restNum, err := strconv.Atoi(currentlyEditing.Name); err == nil {
		if restNum > 0 && restNum <= len(exitOptions) {
			currentlyEditing.Name = exitOptions[restNum-1].Name
		}
	}

	for _, o := range exitOptions {
		if strings.EqualFold(o.Name, currentlyEditing.Name) {
			currentlyEditing.Name = o.Name
			break
		}
	}

	// Load the (possible) existing exit
	currentlyEditing.Exit, currentlyEditing.Exists = room.Exits[currentlyEditing.Name]

	// If they entered a exit name...
	if currentlyEditing.Name != `new` {

		// Does the exit name they entered not exist? Failure!
		if !currentlyEditing.Exists {
			user.SendText("Invalid option selected.")
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		// Since they picked a exit that exists, lets get the question of delete out of the way immediately.
		question := cmdPrompt.Ask(`Delete this exit?`, []string{`yes`, `no`}, `no`)
		if !question.Done {
			return true, nil
		}

		// Delete the exit if that's what they want!
		if question.Response == `yes` {

			delete(room.Exits, currentlyEditing.Name)
			rooms.SaveRoom(*room)

			user.SendText(``)
			user.SendText(fmt.Sprintf(`<ansi fg="exit">%s</ansi> deleted from the room.`, currentlyEditing.Name))
			user.SendText(``)

			user.ClearPrompt()
			return true, nil
		}

	}

	//
	// Name Selection
	//
	{
		// If they are creating a new exit, we don't want that to become a viable exit name, lets empty it
		if currentlyEditing.Name == `new` {
			currentlyEditing.Name = ``
		}

		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Choose a name for this exit:`, []string{currentlyEditing.Name}, currentlyEditing.Name)
		if !question.Done {
			return true, nil
		}
		currentlyEditing.NameNew = question.Response

		// Make sure they aren't using any reserved names.
		if currentlyEditing.NameNew == `quit` || currentlyEditing.NameNew == `new` {
			user.SendText("Invalid new name selected.")
			user.SendText("Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		// Make sure the new name isn't a duplicate
		if currentlyEditing.Name != currentlyEditing.NameNew {
			if _, ok := room.Exits[currentlyEditing.NameNew]; ok {

				user.SendText(`<ansi fg="red">An exit with that name already exists!</ansi>`)
				question.RejectResponse()
				return true, nil

			}
		}

	}

	//
	// Target RoomId
	//
	{
		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`What RoomId will this exit lead to?`, []string{strconv.Itoa(currentlyEditing.Exit.RoomId)}, strconv.Itoa(currentlyEditing.Exit.RoomId))
		if !question.Done {
			return true, nil
		}

		currentlyEditing.Exit.RoomId, _ = strconv.Atoi(question.Response)

		// Make sure they aren't using any reserved names.
		if rooms.LoadRoom(currentlyEditing.Exit.RoomId) == nil {
			user.SendText("Invalid RoomId provided.")
			question.RejectResponse()
			return true, nil
		}

	}

	//
	// Exit message?
	//
	{
		secretExitDefault := `no`
		if currentlyEditing.Exit.Secret {
			secretExitDefault = `yes`
		}

		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Is this a hidden exit?`, []string{`yes`, `no`}, secretExitDefault)
		if !question.Done {
			return true, nil
		}

		currentlyEditing.Exit.Secret = question.Response == `yes`
	}

	//
	// Secret exit?
	//
	{
		secretExitDefault := `no`
		if currentlyEditing.Exit.Secret {
			secretExitDefault = `yes`
		}

		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Is this a hidden exit?`, []string{`yes`, `no`}, secretExitDefault)
		if !question.Done {
			return true, nil
		}

		currentlyEditing.Exit.Secret = question.Response == `yes`
	}

	//
	// Special message when using the exit?
	//
	{
		defaultMessage := currentlyEditing.Exit.ExitMessage
		if defaultMessage == `` {
			defaultMessage = `none`
		}
		// allow them to name/rename the exit.
		question := cmdPrompt.Ask(`Special message when using the exit?`, []string{defaultMessage}, defaultMessage)
		if !question.Done {
			return true, nil
		}

		if question.Response != `none` {
			currentlyEditing.Exit.ExitMessage = question.Response
		}

	}

	//
	// Lock Options
	//
	{
		question := cmdPrompt.Ask(`Will this exit be locked?`, []string{`yes`, `no`}, util.BoolYN(currentlyEditing.Exit.Lock.Difficulty > 0))
		if !question.Done {
			return true, nil
		}

		if question.Response == `yes` {

			defaultDifficultyAnswer := ``
			if currentlyEditing.Exit.Lock.Difficulty > 0 {
				defaultDifficultyAnswer = strconv.Itoa(int(currentlyEditing.Exit.Lock.Difficulty))
			}

			question := cmdPrompt.Ask(`What difficulty will the lock be (2-32)?`, []string{defaultDifficultyAnswer}, defaultDifficultyAnswer)
			if !question.Done {
				return true, nil
			}

			difficultyInt, _ := strconv.Atoi(question.Response)

			// Make sure the provided difficulty is within acceptable range.
			if difficultyInt < 2 || difficultyInt > 32 {
				user.SendText("Difficulty must between 2 and 32, inclusive.")
				question.RejectResponse()
				return true, nil
			}

			currentlyEditing.Exit.Lock.Difficulty = uint8(difficultyInt)

		} else {
			// reset the lock state if there is no lock.
			currentlyEditing.Exit.Lock = gamelock.Lock{}
		}

		if currentlyEditing.Exit.Lock.Difficulty > 0 {
			//
			// Lock Trap Options
			//
			question = cmdPrompt.Ask(`Will this lock have a trap?`, []string{`yes`, `no`}, util.BoolYN(len(currentlyEditing.Exit.Lock.TrapBuffIds) > 0))
			if !question.Done {
				return true, nil
			}

			if question.Response == `yes` {

				selectedBuffList := []int{}
				if cb, ok := cmdPrompt.Recall(`trapBuffs`); ok {
					selectedBuffList = cb.([]int)
				}

				if len(selectedBuffList) == 0 {
					selectedBuffList = append(selectedBuffList, currentlyEditing.Exit.Lock.TrapBuffIds...)
				}

				// Keep track of the state
				cmdPrompt.Store(`trapBuffs`, selectedBuffList)

				selectedBuffLookup := map[int]bool{}
				for _, bId := range selectedBuffList {
					selectedBuffLookup[bId] = true
				}

				buffOptions := []templates.NameDescription{}

				for _, buffId := range buffs.GetAllBuffIds() {
					if b := buffs.GetBuffSpec(buffId); b != nil {

						if b.Name == `empty` {
							continue
						}

						marked := false
						if _, ok := selectedBuffLookup[buffId]; ok {
							marked = true
						}

						buffOptions = append(buffOptions, templates.NameDescription{Id: buffId, Marked: marked, Name: b.Name})
					}
				}

				sort.SliceStable(buffOptions, func(i, j int) bool {
					return buffOptions[i].Name < buffOptions[j].Name
				})

				question := cmdPrompt.Ask(`Select a buff to add to the trap, or nothing to continue:`, []string{}, `0`)
				if !question.Done {
					tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions)
					user.SendText(tplTxt)
					return true, nil
				}

				buffSelected := question.Response

				if buffSelected != `0` {

					buffSelectedInt := 0

					if restNum, err := strconv.Atoi(buffSelected); err == nil {
						if restNum > 0 && restNum <= len(buffOptions) {
							buffSelectedInt = buffOptions[restNum-1].Id.(int)
						}
					}

					if buffSelectedInt == 0 {
						for _, b := range buffOptions {
							if strings.EqualFold(b.Name, buffSelected) {
								buffSelectedInt = b.Id.(int)
								break
							}
						}
					}

					if buffSelectedInt == 0 {

						user.SendText("Invalid selection.")
						question.RejectResponse()

						tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions)
						user.SendText(tplTxt)
						return true, nil
					}

					if _, ok := selectedBuffLookup[buffSelectedInt]; ok {

						delete(selectedBuffLookup, buffSelectedInt)
						for idx, buffId := range selectedBuffList {
							if buffId == buffSelectedInt {
								selectedBuffList = append(selectedBuffList[0:idx], selectedBuffList[idx+1:]...)
								break
							}
						}

					} else {

						selectedBuffList = append(selectedBuffList, buffSelectedInt)
						selectedBuffLookup[buffSelectedInt] = true

					}

					cmdPrompt.Store(`trapBuffs`, selectedBuffList)

					question.RejectResponse()

					for idx, data := range buffOptions {
						_, data.Marked = selectedBuffLookup[data.Id.(int)]
						buffOptions[idx] = data
					}

					tplTxt, _ := templates.Process("tables/numbered-list-doubled", buffOptions)
					user.SendText(tplTxt)
					return true, nil

				}

			}

			if cb, ok := cmdPrompt.Recall(`trapBuffs`); ok {
				currentlyEditing.Exit.Lock.TrapBuffIds = cb.([]int)
			}

			if currentlyEditing.Exit.Lock.RelockInterval == `` {
				currentlyEditing.Exit.Lock.RelockInterval = gamelock.DefaultRelockTime
			}

			question = cmdPrompt.Ask(`How long until it automatically relocks?`, []string{currentlyEditing.Exit.Lock.RelockInterval}, currentlyEditing.Exit.Lock.RelockInterval)
			if !question.Done {
				return true, nil
			}

			currentlyEditing.Exit.Lock.RelockInterval = question.Response

			// If the default time is chosen, can just leave it blank.
			if currentlyEditing.Exit.Lock.RelockInterval == gamelock.DefaultRelockTime {
				currentlyEditing.Exit.Lock.RelockInterval = ``
			}

		}
	}

	//
	// Done editing. Save results
	//
	if currentlyEditing.Name != `` {
		delete(room.Exits, currentlyEditing.Name)
	}

	room.Exits[currentlyEditing.NameNew] = currentlyEditing.Exit
	rooms.SaveRoom(*room)

	user.SendText(``)

	if currentlyEditing.Exit.Lock.Difficulty > 0 {
		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, currentlyEditing.NameNew)
		user.SendText(fmt.Sprintf(`<ansi fg="red">To Create Key -  LockId: <ansi fg="231" bg="5">%s</ansi></ansi>`, lockId))

		seqString := ``
		for _, dir := range util.GetLockSequence(lockId, int(currentlyEditing.Exit.Lock.Difficulty), string(configs.GetServerConfig().Seed)) {
			seqString += string(dir) + " "
		}
		user.SendText(fmt.Sprintf(`<ansi fg="red">To pick lock - Sequence: <ansi fg="green">%s</ansi></ansi>`, seqString))
	}

	user.SendText(``)
	user.SendText(`Changes saved.`)
	user.SendText(``)

	user.ClearPrompt()

	return true, nil
}

func room_Edit_Mutators(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	allRoomMutators := []string{}
	for _, roomMut := range room.Mutators {
		allRoomMutators = append(allRoomMutators, roomMut.MutatorId)
	}

	cmdPrompt, _ := user.StartPrompt(`room edit mutators`, rest)

	selectedMutatorList := []string{}
	if muts, ok := cmdPrompt.Recall(`mutators`); ok {
		selectedMutatorList = muts.([]string)
	} else {
		if len(selectedMutatorList) == 0 {
			selectedMutatorList = append(selectedMutatorList, allRoomMutators...)
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

	question := cmdPrompt.Ask(`Select a mutator to add to the room, or nothing to continue:`, []string{}, `0`)
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
	// Done editing. Save results
	//
	room.Mutators = mutators.MutatorList{}
	for _, mutId := range selectedMutatorList {
		room.Mutators = append(room.Mutators, mutators.Mutator{MutatorId: mutId})
	}
	rooms.SaveRoom(*room)

	user.SendText(``)
	user.SendText(`Changes saved.`)
	user.SendText(``)

	user.ClearPrompt()

	return true, nil
}
