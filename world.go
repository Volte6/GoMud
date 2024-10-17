package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/volte6/mud/badinputtracker"
	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/colorpatterns"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/mobcommands"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/prompt"
	"github.com/volte6/mud/quests"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/usercommands"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

type WorldInput struct {
	FromId    int
	InputText string
	WaitTurns int
}

func (wi WorldInput) Id() int {
	return wi.FromId
}

type World struct {
	connectionPool *connection.ConnectionTracker
	users          *users.ActiveUsers
	worldInput     chan WorldInput
}

func (w *World) GetUsers() *users.ActiveUsers {
	return w.users
}

func (w *World) GetConnectionPool() *connection.ConnectionTracker {
	return w.connectionPool
}

// Send input to the world.
// Just sends via a channel. Will block until read.
func (w *World) Input(i WorldInput) {
	w.worldInput <- i
}

func (w *World) EnterWorld(roomId int, zone string, userId int) {

	user := users.GetByUserId(userId)
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {

		slog.Error("EnterWorld", "error", fmt.Sprintf(`room %d not found`, user.Character.RoomId))

		user.Character.RoomId = 1
		user.Character.Zone = "Frostfang"
		room = rooms.LoadRoom(user.Character.RoomId)
		if room == nil {
			slog.Error("EnterWorld", "error", fmt.Sprintf(`room %d not found`, user.Character.RoomId))
		}
	}

	// TODO HERE
	loginCmds := configs.GetConfig().OnLoginCommands
	if len(loginCmds) > 0 {

		for _, cmd := range loginCmds {

			events.AddToQueue(events.Input{
				UserId:    userId,
				InputText: cmd,
				WaitTurns: -1, // No delay between execution of commands
			})

		}

	}

	// Pu thtme in the room
	rooms.MoveToRoom(userId, roomId, true)
}

func (w *World) LeaveWorld(userId int) {
	user := users.GetByUserId(userId)
	if user == nil {
		return
	}

	room := rooms.LoadRoom(user.Character.RoomId)

	if currentParty := parties.Get(userId); currentParty != nil {
		currentParty.Leave(userId)
	}

	for _, mobInstId := range room.GetMobs(rooms.FindCharmed) {
		if mob := mobs.GetInstance(mobInstId); mob != nil {
			if mob.Character.IsCharmed(userId) {
				mob.Character.Charmed.Expire()
			}
		}
	}

	if _, ok := room.RemovePlayer(userId); ok {
		connectionIds := users.GetConnectionIds(room.GetPlayers())
		tplTxt, _ := templates.Process("player-despawn", user.Character.Name)
		w.connectionPool.SendTo([]byte(tplTxt), connectionIds...)
	}

}

func (w *World) GetAutoComplete(userId int, inputText string) []string {

	suggestions := []string{}

	user := users.GetByUserId(userId)
	if user == nil {
		return suggestions
	}

	// If engaged in a prompt just try and match an option
	if promptInfo := user.GetPrompt(); promptInfo != nil {
		if qInfo := promptInfo.GetNextQuestion(); qInfo != nil {

			if len(qInfo.Options) > 0 {

				for _, opt := range qInfo.Options {

					if inputText == `` {
						suggestions = append(suggestions, opt)
						continue
					}

					s1 := strings.ToLower(opt)
					s2 := strings.ToLower(inputText)
					if s1 != s2 && strings.HasPrefix(s1, s2) {
						suggestions = append(suggestions, s1[len(s2):])
					}
				}

				return suggestions
			}
		}
	}

	if inputText == `` {
		return suggestions
	}

	isAdmin := user.Permission == users.PermissionAdmin
	parts := strings.Split(inputText, ` `)

	// If only one part, probably a command
	if len(parts) < 2 {

		suggestions = append(suggestions, usercommands.GetCmdSuggestions(parts[0], isAdmin)...)

		if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
			for exitName, exitInfo := range room.Exits {
				if exitInfo.Secret {
					continue
				}
				if strings.HasPrefix(strings.ToLower(exitName), strings.ToLower(parts[0])) {
					suggestions = append(suggestions, exitName[len(parts[0]):])
				}
			}
		}
	} else {

		cmd := keywords.TryCommandAlias(parts[0])
		targetName := strings.ToLower(strings.Join(parts[1:], ` `))
		targetNameLen := len(targetName)

		itemList := []items.Item{}
		itemTypeSearch := []items.ItemType{}
		itemSubtypeSearch := []items.ItemSubType{}

		if cmd == `help` {

			suggestions = append(suggestions, usercommands.GetHelpSuggestions(targetName, isAdmin)...)

		} else if cmd == `look` {

			itemList = user.Character.GetAllBackpackItems()

			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				for exitName, exitInfo := range room.Exits {
					if exitInfo.Secret {
						continue
					}
					if strings.HasPrefix(strings.ToLower(exitName), targetName) {
						suggestions = append(suggestions, exitName[targetNameLen:])
					}
				}

				for containerName, _ := range room.Containers {
					if strings.HasPrefix(strings.ToLower(containerName), targetName) {
						suggestions = append(suggestions, containerName[targetNameLen:])
					}
				}
			}

		} else if cmd == `drop` || cmd == `trash` || cmd == `sell` || cmd == `store` || cmd == `inspect` || cmd == `enchant` || cmd == `appraise` || cmd == `give` {

			itemList = user.Character.GetAllBackpackItems()

			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				for exitName, exitInfo := range room.Exits {
					if exitInfo.Secret {
						continue
					}
					if strings.HasPrefix(strings.ToLower(exitName), targetName) {
						suggestions = append(suggestions, exitName[targetNameLen:])
					}
				}

				for containerName, _ := range room.Containers {
					if strings.HasPrefix(strings.ToLower(containerName), targetName) {
						suggestions = append(suggestions, containerName[targetNameLen:])
					}
				}
			}

		} else if cmd == `equip` {

			itemList = user.Character.GetAllBackpackItems()
			itemSubtypeSearch = append(itemSubtypeSearch, items.Wearable)
			itemTypeSearch = append(itemTypeSearch, items.Weapon)

		} else if cmd == `remove` {

			itemList = user.Character.GetAllWornItems()

		} else if cmd == `get` {

			// all items on the floor
			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				itemList = room.GetAllFloorItems(false)
			}

			// Matches for things in containers
			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				if room.Gold > 0 {
					goldName := `gold`
					if strings.HasPrefix(goldName, targetName) {
						suggestions = append(suggestions, goldName[targetNameLen:])
					}
				}
				for containerName, containerInfo := range room.Containers {
					if containerInfo.Lock.IsLocked() {
						continue
					}

					for _, item := range containerInfo.Items {
						iSpec := item.GetSpec()
						if strings.HasPrefix(strings.ToLower(iSpec.Name), targetName) {
							suggestions = append(suggestions, iSpec.Name[targetNameLen:]+` from `+containerName)
						}
					}

					if containerInfo.Gold > 0 {
						goldName := `gold from ` + containerName
						if strings.HasPrefix(goldName, targetName) {
							suggestions = append(suggestions, goldName[targetNameLen:])
						}
					}

				}
			}

		} else if cmd == `eat` {

			itemList = user.Character.GetAllBackpackItems()
			itemSubtypeSearch = append(itemSubtypeSearch, items.Edible)

		} else if cmd == `drink` {

			itemList = user.Character.GetAllBackpackItems()
			itemSubtypeSearch = append(itemSubtypeSearch, items.Drinkable)

		} else if cmd == `use` {

			itemList = user.Character.GetAllBackpackItems()
			itemSubtypeSearch = append(itemSubtypeSearch, items.Usable)

		} else if cmd == `throw` {

			itemList = user.Character.GetAllBackpackItems()
			itemSubtypeSearch = append(itemSubtypeSearch, items.Throwable)

		} else if cmd == `picklock` || cmd == `unlock` || cmd == `lock` {

			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				for exitName, exitInfo := range room.Exits {
					if exitInfo.Secret || !exitInfo.HasLock() {
						continue
					}
					if strings.HasPrefix(strings.ToLower(exitName), targetName) {
						suggestions = append(suggestions, exitName[targetNameLen:])
					}
				}

				for containerName, containerInfo := range room.Containers {
					if containerInfo.HasLock() {
						if strings.HasPrefix(strings.ToLower(containerName), targetName) {
							suggestions = append(suggestions, containerName[targetNameLen:])
						}
					}
				}
			}

		} else if cmd == `attack` || cmd == `consider` {

			// Get all mobs in the room who are not charmed
			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {

				mobNameTracker := map[string]int{}

				for _, mobInstId := range room.GetMobs() {
					if mob := mobs.GetInstance(mobInstId); mob != nil {

						if mob.Character.IsCharmed() && (mob.Character.Aggro == nil || mob.Character.Aggro.UserId != userId) {
							continue
						}

						if targetName == `` {
							suggestions = append(suggestions, mob.Character.Name)
							continue
						}

						if strings.HasPrefix(strings.ToLower(mob.Character.Name), targetName) {
							name := mob.Character.Name[targetNameLen:]

							mobNameTracker[name] = mobNameTracker[name] + 1

							if mobNameTracker[name] > 1 {
								name += `#` + strconv.Itoa(mobNameTracker[name])
							}
							suggestions = append(suggestions, name)

						}
					}
				}

			}
		} else if cmd == `buy` {

			if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
				for _, mobInstId := range room.GetMobs(rooms.FindMerchant) {

					mob := mobs.GetInstance(mobInstId)
					if mob == nil {
						continue
					}

					for _, stockInfo := range mob.Character.Shop.GetInstock() {
						item := items.New(stockInfo.ItemId)
						if item.ItemId > 0 {
							itemList = append(itemList, item)
						}
					}
				}
			}

		} else if cmd == `set` {

			options := []string{
				`description`,
				`prompt`,
				`fprompt`,
				`tinymap`,
			}

			for _, opt := range options {
				if strings.HasPrefix(opt, targetName) {
					suggestions = append(suggestions, opt[len(targetName):])
				}
			}

		} else if cmd == `spawn` {

			if len(inputText) >= len(`spawn item `) && inputText[0:len(`spawn item `)] == `spawn item ` {
				targetName := inputText[len(`spawn item `):]
				for _, itemName := range items.GetAllItemNames() {
					for _, testName := range util.BreakIntoParts(itemName) {
						if strings.HasPrefix(testName, targetName) {
							suggestions = append(suggestions, testName[len(targetName):])
						}
					}
				}
			} else if len(inputText) >= len(`spawn mob `) && inputText[0:len(`spawn mob `)] == `spawn mob ` {
				targetName := inputText[len(`spawn mob `):]
				for _, mobName := range mobs.GetAllMobNames() {
					for _, testName := range util.BreakIntoParts(mobName) {
						if strings.HasPrefix(testName, targetName) {
							suggestions = append(suggestions, testName[len(targetName):])
						}
					}
				}
			} else if len(inputText) >= len(`spawn gold `) && inputText[0:len(`spawn gold `)] == `spawn gold ` {
				suggestions = append(suggestions, "50", "100", "500", "1000", "5000")
			} else {
				options := []string{
					`mob`,
					`gold`,
					`item`,
				}

				for _, opt := range options {
					if strings.HasPrefix(opt, targetName) {
						suggestions = append(suggestions, opt[len(targetName):])
					}
				}
			}

		} else if cmd == `locate` {

			ids := users.GetOnlineUserIds()
			for _, id := range ids {
				if id == user.UserId {
					continue
				}
				if user := users.GetByUserId(id); user != nil {
					if strings.HasPrefix(strings.ToLower(user.Character.Name), targetName) {
						suggestions = append(suggestions, user.Character.Name[targetNameLen:])
					}
				}
			}

		} else if cmd == `cast` {
			for spellName, casts := range user.Character.GetSpells() {
				if casts < 0 {
					continue
				}
				if strings.HasPrefix(spellName, targetName) {
					suggestions = append(suggestions, spellName[len(targetName):])
				}
			}
		}

		itmCt := len(itemList)
		if itmCt > 0 {

			// Keep track of how many times this name occurs to ennumerate the names in suggestions
			// Example: dagger, dagger#2, dagger#3 etc
			bpItemTracker := map[string]int{}

			typeSearchCt := len(itemTypeSearch)
			subtypeSearchCt := len(itemSubtypeSearch)

			for _, item := range itemList {
				iSpec := item.GetSpec()

				skip := false
				if typeSearchCt > 0 || subtypeSearchCt > 0 {
					skip = true

					for i := 0; i < typeSearchCt; i++ {
						if iSpec.Type == itemTypeSearch[i] {
							skip = false
						}
					}

					for i := 0; i < subtypeSearchCt; i++ {
						if iSpec.Subtype == itemSubtypeSearch[i] {
							skip = false
						}
					}

					if skip {
						continue
					}
				}

				if targetName == `` {

					name := iSpec.Name

					bpItemTracker[name] = bpItemTracker[name] + 1

					if bpItemTracker[name] > 1 {
						name += `#` + strconv.Itoa(bpItemTracker[name])
					}
					suggestions = append(suggestions, name)

					continue
				}

				for _, testName := range util.BreakIntoParts(iSpec.Name) {
					if strings.HasPrefix(strings.ToLower(testName), targetName) {
						name := testName[targetNameLen:]

						bpItemTracker[name] = bpItemTracker[name] + 1

						if bpItemTracker[name] > 1 {
							name += `#` + strconv.Itoa(bpItemTracker[name])
						}
						suggestions = append(suggestions, name)
					}
				}
			}

		}

	}
	// Sort by shortest matches first
	sort.Slice(suggestions, func(i, j int) bool {
		return len(suggestions[i]) < len(suggestions[j])
	})

	return suggestions
}

const (
	// Used in GameTickWorker()
	// Used in MaintenanceWorker()
	roomMaintenancePeriod = time.Second * 3  // Every 3 seconds run room maintenance.
	serverStatsLogPeriod  = time.Second * 60 // Every 60 seconds log server stats.
	ansiAliasReloadPeriod = time.Second * 4  // Every 4 seconds reload ansi aliases.
)

func (w *World) GameTickWorker(shutdown chan bool, wg *sync.WaitGroup) {
	wg.Add(1)

	slog.Info("GameTickWorker", "state", "Started")
	defer func() {
		slog.Error("GameTickWorker", "state", "Stopped")
		wg.Done()
	}()

	c := configs.GetConfig()

	messageTimer := time.NewTimer(time.Millisecond)
	turnTimer := time.NewTimer(time.Duration(c.TurnMs) * time.Millisecond)

loop:
	for {
		select {
		case <-shutdown:
			slog.Error(`GameTickWorker`, `action`, `shutdown received`)
			break loop

		case <-messageTimer.C:
			messageTimer.Reset(time.Millisecond)
			w.MessageTick()

		case <-turnTimer.C:
			turnTimer.Reset(time.Duration(c.TurnMs) * time.Millisecond)
			w.TurnTick()

		}
		c = configs.GetConfig()
	}

}

func (w *World) MaintenanceWorker(shutdown chan bool, wg *sync.WaitGroup) {
	wg.Add(1)

	slog.Info("MaintenanceWorker", "state", "Started")
	defer func() {
		slog.Error("MaintenanceWorker", "state", "Stopped")
		wg.Done()
	}()

	roomUpdateTimer := time.NewTimer(roomMaintenancePeriod)
	ansiAliasTimer := time.NewTimer(ansiAliasReloadPeriod)
	//serverStatsLogTimer := time.NewTimer(serverStatsLogPeriod)

loop:
	for {
		select {
		case <-shutdown:
			slog.Error(`MaintenanceWorker`, `action`, `shutdown received`)
			if err := rooms.SaveAllRooms(); err != nil {
				slog.Error("rooms.SaveAllRooms()", "error", err.Error())
			}
			// Save all user data too.
			users.SaveAllUsers()

			break loop

		case <-roomUpdateTimer.C:
			slog.Debug(`MaintenanceWorker`, `action`, `rooms.RoomMaintenance()`)
			rooms.RoomMaintenance(w.connectionPool)
			roomUpdateTimer.Reset(roomMaintenancePeriod)

		case <-ansiAliasTimer.C:
			templates.LoadAliases()
			ansiAliasTimer.Reset(ansiAliasReloadPeriod)

			//case <-serverStatsLogTimer.C:
			//serverStats := util.ServerStats()
			//fmt.Println(templates.AnsiParse(serverStats))
			//serverStatsLogTimer.Reset(serverStatsLogPeriod)

		}
	}

}

func (w *World) InputWorker(shutdown chan bool, wg *sync.WaitGroup) {
	wg.Add(1)

	slog.Info("InputWorker", "state", "Started")
	defer func() {
		slog.Error("InputWorker", "state", "Stopped")
		wg.Done()
	}()

loop:
	for {
		select {
		case <-shutdown:
			slog.Error(`InputWorker`, `action`, `shutdown received`)
			break loop
		case wi := <-w.worldInput:

			events.AddToQueue(events.Input{
				UserId:    wi.FromId,
				InputText: wi.InputText,
				WaitTurns: wi.WaitTurns,
			})

		}
	}
}

func (w *World) processInput(userId int, inputText string) {

	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		slog.Error("User not found", "userId", userId)
		return
	}

	connId := user.ConnectionId()

	var activeQuestion *prompt.Question = nil

	if cmdPrompt := user.GetPrompt(); cmdPrompt != nil {

		if activeQuestion = cmdPrompt.GetNextQuestion(); activeQuestion != nil {

			activeQuestion.Answer(string(inputText))
			inputText = ``

			// set the input buffer to invoke the command prompt it was relevant to
			if cmdPrompt.Command != `` {
				inputText = cmdPrompt.Command + " " + cmdPrompt.Rest
			}
		} else {
			// If a prompt was found, but no pending questions, clear it.
			user.ClearPrompt()
		}

	}

	command := ``
	remains := ``

	var err error
	handled := false

	inputText = strings.TrimSpace(inputText)

	if len(inputText) > 0 {

		// Update their last input
		// Must be actual text, blank space doesn't count.
		user.SetLastInputRound(util.GetRoundCount())

		// Check for macros
		if user.Macros != nil && len(inputText) == 2 {
			if macro, ok := user.Macros[inputText]; ok {
				handled = true
				for waitTime, newCmd := range strings.Split(macro, `;`) {
					if newCmd == `` {
						continue
					}

					events.AddToQueue(events.Input{
						UserId:    userId,
						InputText: newCmd,
						WaitTurns: waitTime,
					})

				}
			}
		}

		if !handled {

			// Lets users use gossip/say shortcuts without a space
			if len(inputText) > 1 {
				if inputText[0] == '`' || inputText[0] == '.' {
					inputText = fmt.Sprintf(`%s %s`, string(inputText[0]), string(inputText[1:]))
				}
			}

			if index := strings.Index(inputText, " "); index != -1 {
				command, remains = strings.ToLower(inputText[0:index]), inputText[index+1:]
			} else {
				command = inputText
			}

			handled, err = usercommands.TryCommand(command, remains, userId)
			if err != nil {
				slog.Error("user-TryCommand", "command", command, "remains", remains, "error", err.Error())
			}
		}

	}

	if !handled {
		if len(command) > 0 {

			badinputtracker.TrackBadCommand(command, remains)

			user.SendText(fmt.Sprintf(`<ansi fg="command">%s</ansi> not recognized. Type <ansi fg="command">help</ansi> for commands.`, command))
			user.Command(`emote @looks a little confused`)
		}
	}

	if worldManager.connectionPool.IsWebsocket(connId) {
		worldManager.connectionPool.SendTo([]byte(user.GetCommandPrompt(true)), connId)
	} else {
		worldManager.connectionPool.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt(true))), connId)
	}

}

func (w *World) processMobInput(mobInstanceId int, inputText string) {
	// No need to select the channel this way

	mob := mobs.GetInstance(mobInstanceId)
	if mob == nil { // Something went wrong. User not found.
		slog.Error("Mob not found", "mobId", mobInstanceId, "where", "processMobInput()")
		return
	}

	command := ""
	remains := ""

	handled := false
	var err error

	if len(inputText) > 0 {

		if index := strings.Index(inputText, " "); index != -1 {
			command, remains = strings.ToLower(inputText[0:index]), inputText[index+1:]
		} else {
			command = inputText
		}

		//slog.Info("World received mob input", "InputText", (inputText))

		handled, err = mobcommands.TryCommand(command, remains, mobInstanceId)
		if err != nil {
			slog.Error("mob-TryCommand", "command", command, "remains", remains, "error", err.Error())
		}

	}

	if !handled {
		if len(command) > 0 {
			mob.Command(fmt.Sprintf(`emote looks a little confused (%s %s).`, command, remains))
		}
	}

}

// Handles sending out queued up messaged to users
func (w *World) MessageTick() {

	//
	// Dispatch GMCP events
	//
	eq := events.GetQueue(events.GMCPOut{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		gmcp, typeOk := e.(events.GMCPOut)
		if !typeOk {
			slog.Error("Event", "Expected Type", "GMCPOut", "Actual Type", e.Type())
			continue
		}

		if gmcp.UserId < 1 {
			continue
		}

		if user := users.GetByUserId(gmcp.UserId); user != nil {
			payload, err := json.Marshal(gmcp.Payload)
			if err != nil {
				slog.Error("Event", "Type", "GMCPOut", "data", gmcp.Payload, "error", err)
				continue
			}
			w.connectionPool.SendTo([]byte(payload), user.ConnectionId())
		}

	}

	//
	// System-wide broadcasts
	//
	eq = events.GetQueue(events.Broadcast{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		broadcast, typeOk := e.(events.Broadcast)
		if !typeOk {
			slog.Error("Event", "Expected Type", "Broadcast", "Actual Type", e.Type())
			continue
		}

		messageColorized := templates.AnsiParse(broadcast.Text)

		if broadcast.SkipLineRefresh {
			w.connectionPool.Broadcast([]byte(messageColorized), []byte(broadcast.Text))
			return
		}

		w.connectionPool.Broadcast(
			[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+messageColorized),
			[]byte(broadcast.Text),
		)
	}

	redrawPrompts := make(map[uint64]string)

	eq = events.GetQueue(events.WebClientCommand{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		cmd, typeOk := e.(events.WebClientCommand)
		if !typeOk {
			slog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
			continue
		}

		if !w.connectionPool.IsWebsocket(cmd.ConnectionId) {
			continue
		}

		w.connectionPool.SendTo([]byte(cmd.Text), cmd.ConnectionId)

	}

	//
	// Outbound text strings
	//
	eq = events.GetQueue(events.Message{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		message, typeOk := e.(events.Message)
		if !typeOk {
			slog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
			continue
		}

		messageColorized := templates.AnsiParse(message.Text)

		slog.Debug("Message{}", "userId", message.UserId, "roomId", message.RoomId, "length", len(messageColorized))

		if message.UserId > 0 {

			if user := users.GetByUserId(message.UserId); user != nil {

				if w.connectionPool.IsWebsocket(user.ConnectionId()) {
					w.connectionPool.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+message.Text), user.ConnectionId())
					if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
						redrawPrompts[user.ConnectionId()] = user.GetCommandPrompt(true)
					}
				} else {
					w.connectionPool.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+messageColorized), user.ConnectionId())
					if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
						redrawPrompts[user.ConnectionId()] = templates.AnsiParse(user.GetCommandPrompt(true))
					}
				}

			}
		}

		if message.RoomId > 0 {

			room := rooms.LoadRoom(message.RoomId)
			if room == nil {
				continue
			}

			for _, userId := range room.GetPlayers() {
				skip := false

				if message.UserId == userId {
					continue
				}

				exLen := len(message.ExcludeUserIds)
				if exLen > 0 {
					for _, excludeId := range message.ExcludeUserIds {
						if excludeId == userId {
							skip = true
							break
						}
					}
				}

				if skip {
					continue
				}

				if user := users.GetByUserId(userId); user != nil {

					// If this is a quiet message, make sure the player can hear it
					if message.IsQuiet {
						if !user.Character.HasBuffFlag(buffs.SuperHearing) {
							continue
						}
					}

					if w.connectionPool.IsWebsocket(user.ConnectionId()) {
						w.connectionPool.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+message.Text), user.ConnectionId())
						if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
							redrawPrompts[user.ConnectionId()] = user.GetCommandPrompt(true)
						}
					} else {
						w.connectionPool.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+messageColorized), user.ConnectionId())
						if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
							redrawPrompts[user.ConnectionId()] = templates.AnsiParse(user.GetCommandPrompt(true))
						}

					}

				}
			}

		}

	}

	for connectionId, prompt := range redrawPrompts {
		w.connectionPool.SendTo([]byte(prompt), connectionId)
	}
}

// Turns are much finer resolution than rounds...
// Many turns occur int he time a round does.
// Discrete actions are processed on the turn level
func (w *World) TurnTick() {

	// Grab the current config
	c := configs.GetConfig()

	turnCt := util.IncrementTurnCount()

	//
	// Cleanup any zombies
	//

	expTurns := (uint64(c.ZombieSeconds) * uint64(c.TurnsPerSecond()))

	if expTurns < turnCt {
		expZombies := users.GetExpiredZombies(turnCt - expTurns)
		if len(expZombies) > 0 {

			connIds := users.GetConnectionIds(expZombies)

			for _, userId := range expZombies {
				worldManager.LeaveWorld(userId)
				users.RemoveZombieUser(userId)
			}
			for _, connId := range connIds {
				if err := users.LogOutUserByConnectionId(connId); err != nil {
					slog.Error("Log Out Error", "connectionId", connId, "error", err)
				}
			}

		}
	}

	if turnCt%uint64(c.TurnsPerAutoSave()) == 0 {
		tStart := time.Now()

		events.AddToQueue(events.Broadcast{
			Text: `Saving users...`,
		})

		users.SaveAllUsers()

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		events.AddToQueue(events.Broadcast{
			Text: `Saving rooms...`,
		})

		rooms.SaveAllRooms()

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		util.TrackTime(`Save Game State`, time.Since(tStart).Seconds())
	}

	tStart := time.Now()
	var eq *events.Queue

	//
	// Handle Input Queue
	//
	alreadyProcessed := make(map[int]struct{}) // Keep track of players who already had a command this turn
	eq = events.GetQueue(events.Input{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		input, typeOk := e.(events.Input)
		if !typeOk {
			slog.Error("Event", "Expected Type", "Input", "Actual Type", e.Type())
			continue
		}

		//slog.Debug(`Event`, `type`, input.Type(), `UserId`, input.UserId, `MobInstanceId`, input.MobInstanceId, `WaitTurns`, input.WaitTurns, `InputText`, input.InputText)

		if input.MobInstanceId > 0 {
			if input.WaitTurns < 1 {
				w.processMobInput(input.MobInstanceId, input.InputText)
			} else {
				input.WaitTurns--
				events.Requeue(input)
			}
			continue
		}

		if input.WaitTurns < 0 { // -1 and below, process immediately and don't count towards limit
			w.processInput(input.UserId, input.InputText)
			continue
		}

		if _, ok := alreadyProcessed[input.UserId]; ok {
			events.Requeue(input)
			continue
		}

		if input.WaitTurns == 0 { // 0 means process immediately but wait another turn before processing another from this user
			w.processInput(input.UserId, input.InputText)
			alreadyProcessed[input.UserId] = struct{}{}
		} else {
			input.WaitTurns--
			events.Requeue(input)
		}

	}

	//
	// Handle RoomAction Queue
	//
	eq = events.GetQueue(events.RoomAction{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		action, typeOk := e.(events.RoomAction)
		if !typeOk {
			slog.Error("Event", "Expected Type", "RoomAction", "Actual Type", e.Type())
			continue
		}

		//slog.Debug(`Event`, `type`, action.Type(), `RoomId`, action.RoomId, `SourceUserId`, action.SourceUserId, `SourceMobId`, action.SourceMobId, `WaitTurns`, action.WaitTurns, `Action`, action.Action)

		if action.WaitTurns > 0 {

			if action.WaitTurns%c.TurnsPerRound() == 0 {
				// Get the parts of the command
				parts := strings.SplitN(action.Action, ` `, 3)
				if parts[0] == `detonate` {
					// Make sure the room exists
					room := rooms.LoadRoom(action.RoomId)
					if room == nil {
						continue
					}

					var itemName string

					if len(parts) > 2 {
						itemName = parts[2]
					} else {
						itemName = parts[1]
					}

					itm, found := room.FindOnFloor(itemName, false)
					if !found {
						continue
					}

					room.SendText(fmt.Sprintf(`The <ansi fg="itemname">%s</ansi> looks like it's about to explode...`, itm.DisplayName()))
				}

			}

			action.WaitTurns--
			events.Requeue(action)
			continue
		}

		// Make sure the room exists
		room := rooms.LoadRoom(action.RoomId)
		if room == nil {
			continue
		}

		if rooms.EffectType(action.Action) == rooms.Wildfire {

			if room.AddEffect(rooms.Wildfire) {
				room.SendText(colorpatterns.ApplyColorPattern(`A wildfire burns through the area!`, `flame`, colorpatterns.Stretch))
				room.SendTextToExits(`You notice a `+colorpatterns.ApplyColorPattern(`wildfire`, `flame`, colorpatterns.Stretch)+` start!`, false)
			}

		}

		// Get the parts of the command
		parts := strings.SplitN(action.Action, ` `, 3)

		// Is it a detonation?
		// Possible formats:
		// donate [#mobId|@userId] !itemId:uid
		// TODO: Refactor this into a scripted event/function
		if parts[0] == `detonate` {

			// Detonate can't be the only information
			if len(parts) < 2 {
				continue
			}

			var itemName string
			var targetName string

			if len(parts) > 2 {
				targetName = parts[1]
				itemName = parts[2]
			} else {
				itemName = parts[1]
			}

			itm, found := room.FindOnFloor(itemName, false)
			if !found {
				continue
			}

			iSpec := itm.GetSpec()
			if iSpec.Type != items.Grenade {
				continue
			}

			room.RemoveItem(itm, false)

			room.SendText(`<ansi fg="red">--- --- --- --- --- --- --- --- --- --- --- ---</ansi>`)
			room.SendText(fmt.Sprintf(`The <ansi fg="itemname">%s</ansi> <ansi fg="red">EXPLODES</ansi>!`, itm.DisplayName()))
			room.SendText(`<ansi fg="red">--- --- --- --- --- --- --- --- --- --- --- ---</ansi>`)

			room.SendTextToExits(`You hear a large <ansi fg="red">!!!EXPLOSION!!!</ansi>`, false)

			if len(iSpec.BuffIds) == 0 {
				continue
			}

			hitMobs := true
			hitPlayers := true

			targetPlayerId, targetMobId := room.FindByName(targetName)

			if targetPlayerId > 0 {
				hitMobs = false
			}

			if targetMobId > 0 {
				hitPlayers = false
			}

			events.Requeue(events.RoomAction{
				RoomId: room.RoomId,
				Action: string(rooms.Wildfire),
			})

			if hitPlayers {
				for _, uid := range room.GetPlayers() {
					for _, buffId := range iSpec.BuffIds {
						events.AddToQueue(events.Buff{
							UserId:        uid,
							MobInstanceId: 0,
							BuffId:        buffId,
						})
					}
				}
			}

			if !hitMobs {
				continue
			}

			for _, mid := range room.GetMobs() {

				for _, buffId := range iSpec.BuffIds {
					events.AddToQueue(events.Buff{
						UserId:        0,
						MobInstanceId: mid,
						BuffId:        buffId,
					})
				}

				if action.SourceUserId == 0 {
					continue
				}

				sourceUser := users.GetByUserId(action.SourceUserId)
				if sourceUser == nil {
					continue
				}

				mob := mobs.GetInstance(mid)
				if mob == nil {
					continue
				}

				mob.DamageTaken[sourceUser.UserId] = 0 // Take note that the player did damage this mob.

				if sourceUser.Character.RoomId == mob.Character.RoomId {
					// Mobs get aggro when attacked
					if mob.Character.Aggro == nil {
						mob.PreventIdle = true

						mob.Command(fmt.Sprintf("attack %s", sourceUser.ShorthandId()))

					}
				} else {

					var foundExitName string

					// Look for them nearby and go to them
					for exitName, exitInfo := range room.Exits {
						if exitInfo.RoomId == sourceUser.Character.RoomId {
							foundExitName = exitName
							break
						}
					}

					if foundExitName == `` {
						// Look for them nearby and go to them
						for exitName, exitInfo := range room.ExitsTemp {
							if exitInfo.RoomId == sourceUser.Character.RoomId {

								mob.Command(fmt.Sprintf("go %s", exitName))
								mob.Command(fmt.Sprintf("attack %s", sourceUser.ShorthandId()))

								break
							}
						}
					}

					if foundExitName != `` {

						mob.Command(fmt.Sprintf("go %s", foundExitName))
						mob.Command(fmt.Sprintf("attack %s", sourceUser.ShorthandId()))

					}
				}

			}

		}

	}

	//
	// Handle Buff Queue
	//
	eq = events.GetQueue(events.Buff{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		buff, typeOk := e.(events.Buff)
		if !typeOk {
			slog.Error("Event", "Expected Type", "Buff", "Actual Type", e.Type())
			continue
		}

		slog.Debug(`Event`, `type`, buff.Type(), `UserId`, buff.UserId, `MobInstanceId`, buff.MobInstanceId, `BuffId`, buff.BuffId)

		buffInfo := buffs.GetBuffSpec(buff.BuffId)
		if buffInfo == nil {
			continue
		}

		var targetChar *characters.Character

		if buff.MobInstanceId > 0 {
			buffMob := mobs.GetInstance(buff.MobInstanceId)
			if buffMob == nil {
				continue
			}
			targetChar = &buffMob.Character
		} else {
			buffUser := users.GetByUserId(buff.UserId)
			if buffUser == nil {
				continue
			}
			targetChar = buffUser.Character
		}

		if buff.BuffId < 0 {
			targetChar.RemoveBuff(buffInfo.BuffId * -1)
			continue
		}

		// Apply the buff
		targetChar.AddBuff(buff.BuffId, false)

		//
		// Fire onStart for buff script
		//
		if _, err := scripting.TryBuffScriptEvent(`onStart`, buff.UserId, buff.MobInstanceId, buff.BuffId); err == nil {
			targetChar.TrackBuffStarted(buff.BuffId)
		}

		//
		// If the buff calls for an immediate triggering
		//
		if buffInfo.TriggerNow {
			scripting.TryBuffScriptEvent(`onTrigger`, buff.UserId, buff.MobInstanceId, buff.BuffId)

			if buff.MobInstanceId > 0 && targetChar.Health <= 0 {
				// Mob died
				events.AddToQueue(events.Input{
					MobInstanceId: buff.MobInstanceId,
					InputText:     `suicide`,
				})
			}
		}

	}

	//
	// Handle Quest Queue
	//
	eq = events.GetQueue(events.Quest{})
	for eq.Len() > 0 {

		e := eq.Poll().(events.Event)

		quest, typeOk := e.(events.Quest)
		if !typeOk {
			slog.Error("Event", "Expected Type", "Quest", "Actual Type", e.Type())
			continue
		}

		slog.Debug(`Event`, `type`, quest.Type(), `UserId`, quest.UserId, `QuestToken`, quest.QuestToken)

		// Give them a token
		remove := false
		if quest.QuestToken[0:1] == `-` {
			remove = true
			quest.QuestToken = quest.QuestToken[1:]
		}

		questInfo := quests.GetQuest(quest.QuestToken)
		if questInfo == nil {
			continue
		}

		questUser := users.GetByUserId(quest.UserId)
		if questUser == nil {
			continue
		}

		if remove {
			questUser.Character.ClearQuestToken(quest.QuestToken)
			continue
		}
		// This only succees if the user doesn't have the quest yet or the quest is a later step of one they've started
		if !questUser.Character.GiveQuestToken(quest.QuestToken) {
			continue
		}

		_, stepName := quests.TokenToParts(quest.QuestToken)
		if stepName == `start` {
			if !questInfo.Secret {
				questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You have been given a new quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name))
				questUser.SendText(questUpTxt)
			}
		} else if stepName == `end` {

			if !questInfo.Secret {
				questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You have completed the quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name))
				questUser.SendText(questUpTxt)
			}

			// Message to player?
			if len(questInfo.Rewards.PlayerMessage) > 0 {
				questUser.SendText(questInfo.Rewards.PlayerMessage)
			}
			// Message to room?
			if len(questInfo.Rewards.RoomMessage) > 0 {
				if room := rooms.LoadRoom(questUser.Character.RoomId); room != nil {
					room.SendText(questInfo.Rewards.RoomMessage, questUser.UserId)
				}
			}
			// New quest to start?
			if len(questInfo.Rewards.QuestId) > 0 {

				events.AddToQueue(events.Quest{
					UserId:     questUser.UserId,
					QuestToken: questInfo.Rewards.QuestId,
				})

			}
			// Gold reward?
			if questInfo.Rewards.Gold > 0 {
				questUser.SendText(fmt.Sprintf(`You receive <ansi fg="gold">%d gold</ansi>!`, questInfo.Rewards.Gold))
				questUser.Character.Gold += questInfo.Rewards.Gold
			}
			// Item reward?
			if questInfo.Rewards.ItemId > 0 {
				newItm := items.New(questInfo.Rewards.ItemId)
				questUser.SendText(fmt.Sprintf(`You receive <ansi fg="itemname">%s</ansi>!`, newItm.NameSimple()))
				questUser.Character.StoreItem(newItm)

				iSpec := newItm.GetSpec()
				if iSpec.QuestToken != `` {

					events.AddToQueue(events.Quest{
						UserId:     questUser.UserId,
						QuestToken: iSpec.QuestToken,
					})

				}
			}
			// Buff reward?
			if questInfo.Rewards.BuffId > 0 {

				events.AddToQueue(events.Buff{
					UserId:        questUser.UserId,
					MobInstanceId: 0,
					BuffId:        questInfo.Rewards.BuffId,
				})

			}
			// Experience reward?
			if questInfo.Rewards.Experience > 0 {

				grantXP, xpScale := questUser.Character.GrantXP(questInfo.Rewards.Experience)

				xpMsgExtra := ``
				if xpScale != 100 {
					xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
				}

				questUser.SendText(fmt.Sprintf(`You receive <ansi fg="experience">%d experience points</ansi>%s!`, grantXP, xpMsgExtra))
			}
			// Skill reward?
			if questInfo.Rewards.SkillInfo != `` {
				details := strings.Split(questInfo.Rewards.SkillInfo, `:`)
				if len(details) > 1 {
					skillName := strings.ToLower(details[0])
					skillLevel, _ := strconv.Atoi(details[1])
					currentLevel := questUser.Character.GetSkillLevel(skills.SkillTag(skillName))

					if currentLevel < skillLevel {
						newLevel := questUser.Character.TrainSkill(skillName, skillLevel)

						skillData := struct {
							SkillName  string
							SkillLevel int
						}{
							SkillName:  skillName,
							SkillLevel: newLevel,
						}
						skillUpTxt, _ := templates.Process("character/skillup", skillData)
						questUser.SendText(skillUpTxt)
					}

				}
			}
			// Move them to another room/area?
			if questInfo.Rewards.RoomId > 0 {
				questUser.SendText(`You are suddenly moved to a new place!`)

				if room := rooms.LoadRoom(questUser.Character.RoomId); room != nil {
					room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is suddenly moved to a new place!`, questUser.Character.Name), questUser.UserId)
				}

				rooms.MoveToRoom(questUser.UserId, questInfo.Rewards.RoomId)
			}
		} else {
			if !questInfo.Secret {
				questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You've made progress on the quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name))
				questUser.SendText(questUpTxt)
			}
		}

	}

	//
	// Prune all buffs that have expired.
	//
	w.PruneBuffs()

	//
	// Update movement points for each player
	// TODO: Optimize this to avoid re-loops through users
	//
	for _, uId := range users.GetOnlineUserIds() {
		if user := users.GetByUserId(uId); user != nil {
			user.Character.ActionPoints += 1
			if user.Character.ActionPoints > user.Character.ActionPointsMax.Value {
				user.Character.ActionPoints = user.Character.ActionPointsMax.Value
			}
		}
	}

	if turnCt%uint64(c.TurnsPerSecond()) == 0 {
		w.CheckForLevelUps()

		// TODO: Move this elsewhere
		// Testing concept, later will be replaced with a `mprompt` (modalprompt)
		for _, uId := range users.GetOnlineUserIds() {
			if user := users.GetByUserId(uId); user != nil {

				//if user.GetConfigOption(`mprompt`) != nil {
				events.AddToQueue(events.WebClientCommand{
					ConnectionId: user.ConnectionId(),
					Text:         "MODALADD:mprompt (testing)=" + user.GetCommandPrompt(false, `mprompt`),
				})
				//}
			}
		}
	}

	//
	// End processing of buffs
	//

	util.TrackTime(`World::TurnTick()`, time.Since(tStart).Seconds())

	// After a full round of turns, we can do a round tick.
	if turnCt%uint64(c.TurnsPerRound()) == 0 {
		w.roundTick()
	}

}

// Force disconnect a user (Makes them a zombie)
func (w *World) Kick(userId int) {

	user := users.GetByUserId(userId)
	if user == nil {
		return
	}
	users.SetZombieUser(userId)
	w.connectionPool.Kick(user.ConnectionId())
}

func NewWorld(osSignalChan chan os.Signal) *World {

	w := &World{
		connectionPool: connection.GetPool(),
		users:          users.NewUserManager(),
		worldInput:     make(chan WorldInput),
	}

	w.connectionPool.SetShutdownChan(osSignalChan)

	return w
}
