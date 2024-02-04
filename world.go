package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/mobcommands"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/parties"
	"github.com/volte6/mud/prompt"
	"github.com/volte6/mud/quests"
	"github.com/volte6/mud/rooms"
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

type BuffApply struct {
	ToId   int
	BuffId int
}

func (ba BuffApply) Id() int {
	return ba.ToId
}

type QuestApply struct {
	ToId       int
	QuestToken string
}

func (qa QuestApply) Id() int {
	return qa.ToId
}

type RoomAction struct {
	RoomId       int
	SourceUserId int
	SourceMobId  int
	Action       string
}

func (ra RoomAction) Id() int {
	return ra.RoomId
}

type World struct {
	connectionPool *connection.ConnectionTracker
	users          *users.ActiveUsers
	worldInput     chan WorldInput
	userInputQueue util.LimitQueue[WorldInput]
	mobInputQueue  util.LimitQueue[WorldInput]

	userBuffQueue util.LimitQueue[BuffApply]
	mobBuffQueue  util.LimitQueue[BuffApply]

	userQuestQueue util.LimitQueue[QuestApply]

	roomActionQueue util.LimitQueue[RoomAction]
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
			w.QueueCommand(userId, 0, cmd)
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

		prompt.Clear(userId)
	}

}

func (w *World) GetAutoComplete(userId int, inputText string) []string {

	suggestions := []string{}

	if inputText == `` {
		return suggestions
	}

	user := users.GetByUserId(userId)
	if user == nil {
		return suggestions
	}

	// If engaged in a prompt just try and match an option
	if promptInfo := prompt.Get(userId); promptInfo != nil {
		if qInfo := promptInfo.GetNextQuestion(); qInfo != nil {

			if len(qInfo.Options) > 0 {
				if len(inputText) > 0 {
					for _, opt := range qInfo.Options {
						s1 := strings.ToLower(opt)
						s2 := strings.ToLower(inputText)
						if s1 != s2 && strings.HasPrefix(s1, s2) {
							suggestions = append(suggestions, s1[len(s2):])
						}
					}
				}

				return suggestions
			}
		}
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

		} else if cmd == `look` || cmd == `drop` || cmd == `trash` || cmd == `sell` || cmd == `store` || cmd == `inspect` || cmd == `enchant` || cmd == `appraise` {

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
					if exitInfo.Secret || !exitInfo.Lock.IsLockable() {
						continue
					}
					if strings.HasPrefix(strings.ToLower(exitName), targetName) {
						suggestions = append(suggestions, exitName[targetNameLen:])
					}
				}

				for containerName, containerInfo := range room.Containers {
					if containerInfo.Lock.IsLockable() {
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

					for itemId := range mob.ShopStock {
						item := items.New(itemId)
						if item.ItemId > 0 {
							itemList = append(itemList, item)
						}
					}
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
					suggestions = append(suggestions, iSpec.Name)
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

	turnTimer := time.NewTimer(time.Duration(c.TurnMilliseconds) * time.Millisecond)

loop:
	for {
		select {
		case <-shutdown:
			slog.Error(`GameTickWorker`, `action`, `shutdown received`)
			break loop

		case <-turnTimer.C:
			turnTimer.Reset(time.Duration(c.TurnMilliseconds) * time.Millisecond)
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
			w.userInputQueue.Push(wi)
		}
	}
}

func (w *World) processInput(wi WorldInput) {
	// No need to select the channel this way
	//for wi := range w.worldInput {

	user := users.GetByUserId(wi.FromId)
	if user == nil { // Something went wrong. User not found.
		slog.Error("User not found", "userId", wi.FromId)
		return
	}

	connId := user.ConnectionId()

	var activeQuestion *prompt.Question = nil

	if cmdPrompt := prompt.Get(wi.FromId); cmdPrompt != nil {

		if activeQuestion = cmdPrompt.GetNextQuestion(); activeQuestion != nil {

			activeQuestion.Answer(string(wi.InputText))
			wi.InputText = ``

			// set the input buffer to invoke the command prompt it was relevant to
			if cmdPrompt.Command != `` {
				wi.InputText = cmdPrompt.Command + " " + cmdPrompt.Rest
			}
		} else {
			// If a prompt was found, but no pending questions, clear it.
			prompt.Clear(wi.FromId)
		}

	}

	for {
		command := ``
		remains := ``

		commandResponse := usercommands.NewUserCommandResponse(wi.FromId)
		var err error

		wi.InputText = strings.TrimSpace(wi.InputText)

		if len(wi.InputText) > 0 {

			// Check for macros
			if user.Macros != nil && len(wi.InputText) == 2 {
				if macro, ok := user.Macros[wi.InputText]; ok {
					commandResponse.Handled = true
					for _, newCmd := range strings.Split(macro, `;`) {
						if newCmd == `` {
							continue
						}
						w.userInputQueue.Push(WorldInput{
							FromId:    wi.FromId,
							InputText: newCmd,
						})
					}
				}
			}

			if !commandResponse.Handled {

				// Lets users use gossip/say shortcuts without a space
				if len(wi.InputText) > 1 {
					if wi.InputText[0] == '`' || wi.InputText[0] == '.' {
						wi.InputText = fmt.Sprintf(`%s %s`, string(wi.InputText[0]), string(wi.InputText[1:]))
					}
				}

				if index := strings.Index(wi.InputText, " "); index != -1 {
					command, remains = strings.ToLower(wi.InputText[0:index]), wi.InputText[index+1:]
				} else {
					command = wi.InputText
				}

				slog.Info("World received input", "InputText", (wi.InputText))
				commandResponse, err = usercommands.TryCommand(command, remains, wi.FromId, w)
				if err != nil {
					slog.Error("user-TryCommand", "command", command, "remains", remains, "error", err.Error())
				}
			}

		}

		if !commandResponse.Handled {
			if len(command) > 0 {
				commandResponse.SendUserMessage(wi.FromId,
					fmt.Sprintf(`<ansi fg="command">%s</ansi> not recognized. Type <ansi fg="command">help</ansi> for commands.`, command),
					true)
				commandResponse.SendRoomMessage(user.Character.RoomId,
					fmt.Sprintf(`<ansi fg="username">%s</ansi> looks a little confused.`, user.Character.Name),
					true,
					wi.FromId)
			}
		}

		if commandResponse.Pending() {
			w.DispatchMessages(commandResponse)
		}

		if len(commandResponse.CommandQueue) > 0 {
			for _, cmd := range commandResponse.CommandQueue {

				if cmd.UserId > 0 {
					w.userInputQueue.Push(WorldInput{
						FromId:    cmd.UserId,
						InputText: cmd.Command,
					})
				} else if cmd.MobInstanceId > 0 {
					w.mobInputQueue.Push(WorldInput{
						FromId:    cmd.MobInstanceId,
						InputText: cmd.Command,
					})
				}
			}
			commandResponse.CommandQueue = commandResponse.CommandQueue[:0]
		}

		// Load up any forced commands
		if len(commandResponse.NextCommand) > 0 {
			wi.InputText = commandResponse.NextCommand
			continue
		}

		break
	}

	worldManager.GetConnectionPool().SendTo([]byte(templates.AnsiParse(user.GetPrompt(true))), connId)

}

func (w *World) processMobInput(wi WorldInput) {
	// No need to select the channel this way

	mob := mobs.GetInstance(wi.FromId)
	if mob == nil { // Something went wrong. User not found.
		slog.Error("Mob not found", "mobId", wi.FromId, "where", "processMobInput()")
		return
	}

	for {
		command := ""
		remains := ""

		commandResponse := mobcommands.NewMobCommandResponse(wi.FromId)
		var err error

		if len(wi.InputText) > 0 {

			if index := strings.Index(wi.InputText, " "); index != -1 {
				command, remains = strings.ToLower(wi.InputText[0:index]), wi.InputText[index+1:]
			} else {
				command = wi.InputText
			}

			//slog.Info("World received mob input", "InputText", (wi.InputText))

			commandResponse, err = mobcommands.TryCommand(command, remains, wi.FromId, w)
			if err != nil {
				slog.Error("mob-TryCommand", "command", command, "remains", remains, "error", err.Error())
			}

		}

		if !commandResponse.Handled {
			if len(command) > 0 {
				commandResponse.SendRoomMessage(mob.Character.RoomId,
					fmt.Sprintf(`<ansi fg="mobname">%s</ansi> looks a little confused (%s %s).`, mob.Character.Name, command, remains),
					true)
			}
		}

		if commandResponse.Pending() {
			w.DispatchMessages(commandResponse)
		}

		if len(commandResponse.CommandQueue) > 0 {
			for _, cmd := range commandResponse.CommandQueue {

				if cmd.UserId > 0 {
					w.userInputQueue.Push(WorldInput{
						FromId:    cmd.UserId,
						InputText: cmd.Command,
					})
				} else if cmd.MobInstanceId > 0 {
					w.mobInputQueue.Push(WorldInput{
						FromId:    cmd.MobInstanceId,
						InputText: cmd.Command,
					})
				}
			}
			commandResponse.CommandQueue = commandResponse.CommandQueue[:0]
		}

		// Load up any forced commands
		if len(commandResponse.NextCommand) > 0 {
			wi.InputText = commandResponse.NextCommand
			continue
		}

		break
	}

}

func (w *World) QueueBuff(userId int, mobId int, buffId int) {

	newInput := BuffApply{
		BuffId: buffId,
	}

	if userId > 0 {
		newInput.ToId = userId
		w.userBuffQueue.Push(newInput)
	}
	if mobId > 0 {
		newInput.ToId = mobId
		w.mobBuffQueue.Push(newInput)
	}

}

func (w *World) QueueRoomAction(roomId int, sourceUserId int, sourceMobId int, action string) {

	newInput := RoomAction{
		RoomId:       roomId,
		SourceUserId: sourceUserId,
		SourceMobId:  sourceMobId,
		Action:       action,
	}

	w.roomActionQueue.Push(newInput)

}

func (w *World) QueueQuest(userId int, questToken string) {

	newInput := QuestApply{
		ToId:       userId,
		QuestToken: questToken,
	}

	w.userQuestQueue.Push(newInput)

}

func (w *World) QueueCommand(userId int, mobId int, cmd string, waitTurns ...int) {

	turnsToWait := 0
	if len(waitTurns) > 0 {
		turnsToWait = waitTurns[0]
	}

	newInput := WorldInput{
		InputText: cmd,
		WaitTurns: turnsToWait,
	}

	if userId > 0 {
		newInput.FromId = userId
		w.userInputQueue.Push(newInput)
	}
	if mobId > 0 {
		newInput.FromId = mobId
		w.mobInputQueue.Push(newInput)
	}

}

func (w *World) Broadcast(msg string, skipLineRefresh ...bool) {

	if len(skipLineRefresh) > 0 && skipLineRefresh[0] {
		w.connectionPool.Broadcast([]byte(msg))
		return
	}

	msg = term.AnsiMoveCursorColumn.String() +
		term.AnsiEraseLine.String() +
		msg

	w.connectionPool.Broadcast([]byte(msg))
}

// Optionally capture any user output and return it.
func (w *World) DispatchMessages(u util.MessageQueue) {

	redrawPrompts := make(map[uint64]string)

	for {

		message, err := u.GetNextMessage()
		if err != nil {
			break
		}

		message.Msg = templates.AnsiParse(message.Msg)

		switch message.MsgType {

		case util.MsgUser:
			if user := users.GetByUserId(message.UserId); user != nil {
				message.Msg = term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + message.Msg
				w.connectionPool.SendTo([]byte(message.Msg), user.ConnectionId())
				if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
					redrawPrompts[user.ConnectionId()] = user.GetPrompt(true)
				}
			}

		case util.MsgRoom:

			if message.RoomId == 0 {
				w.connectionPool.Broadcast([]byte(message.Msg))
				break
			}

			if room := rooms.LoadRoom(message.RoomId); err == nil {
				for _, userId := range room.GetPlayers() {
					skip := false

					if u.UserId > 0 && u.UserId == userId {
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
						message.Msg = term.AnsiMoveCursorColumn.String() + term.AnsiEraseLine.String() + message.Msg
						w.connectionPool.SendTo([]byte(message.Msg), user.ConnectionId())
						if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
							redrawPrompts[user.ConnectionId()] = user.GetPrompt(true)
						}
					}
				}
			}
		}

	}

	for connectionId, prompt := range redrawPrompts {
		prompt = templates.AnsiParse(prompt)
		w.connectionPool.SendTo([]byte(prompt), connectionId)
	}

}

func (w *World) GetSettings(userId int) (connection.ClientSettings, error) {

	for _, cid := range users.GetConnectionIds([]int{userId}) {
		if cd, err := w.connectionPool.Get(cid); err == nil {
			return cd.GetSettings(), nil
		}
	}

	return connection.ClientSettings{}, errors.New(`user settings not found`)
}

// Turns are much finer resolution than rounds...
// Many turns occur int he time a round does.
// Discrete actions are processed on the turn level
func (w *World) TurnTick() {

	// Grab the current config
	c := configs.GetConfig()

	turnCt := util.IncrementTurnCount()

	if turnCt%uint64(c.TurnsPerAutoSave()) == 0 {
		tStart := time.Now()

		w.Broadcast(`Saving users...`)
		users.SaveAllUsers()
		w.Broadcast(`Done.`+term.CRLFStr, true)

		w.Broadcast(`Saving rooms...`)
		rooms.SaveAllRooms()
		w.Broadcast(`Done.`+term.CRLFStr, true)

		util.TrackTime(`Save Game State`, time.Since(tStart).Seconds())
	}

	tStart := time.Now()

	requeueList := []WorldInput{}

	// Process any pending inputs by mobs
	for w.mobInputQueue.Len(0) > 0 {
		if wi, ok := w.mobInputQueue.Pop(); ok {
			if wi.WaitTurns < 1 {
				w.processMobInput(wi)
			} else {
				wi.WaitTurns--
				requeueList = append(requeueList, wi)
			}
		}
	}

	if len(requeueList) > 0 {
		slog.Debug(`Mob Requeuing`, `count`, len(requeueList), "inputs", requeueList)
		//fmt.Println(`Requeueing`, len(requeueList), `inputs`, requeueList)
		for _, wi := range requeueList {
			w.mobInputQueue.Push(wi)
		}
	}

	requeueList = requeueList[:0] // Empty the list

	alreadyProcessed := make(map[int]struct{}) // Keep track of players who already had a command this turn
	// Process any pending inputs by players
	for w.userInputQueue.Len(0) > 0 {
		if wi, ok := w.userInputQueue.Pop(); ok {

			if _, ok := alreadyProcessed[wi.FromId]; ok {
				requeueList = append(requeueList, wi)
				continue
			}
			if wi.WaitTurns < 1 {
				w.processInput(wi)
				alreadyProcessed[wi.FromId] = struct{}{}
			} else {
				wi.WaitTurns--
				requeueList = append(requeueList, wi)
			}
		}
	}

	if len(requeueList) > 0 {
		slog.Info(`Usr Requeuing`, `count`, len(requeueList))
		//fmt.Println(`Requeueing`, len(requeueList), `inputs`, requeueList)
		for _, wi := range requeueList {
			w.userInputQueue.Push(wi)
		}
	}

	//
	// The follow section handles queued up buffs
	// They get processed in "TICK" time which is much faster than "ROUND" time
	messageQueue := util.NewMessageQueue(0, 0)

	// Process any pending buffs on mobs
	for w.roomActionQueue.Len(0) > 0 {
		if actionRequest, ok := w.roomActionQueue.Pop(); ok {
			if room := rooms.LoadRoom(actionRequest.RoomId); room != nil {

				// Get the parts of the command
				parts := strings.SplitN(actionRequest.Action, ` `, 3)

				// Is it a detonation?
				// Possible formats:
				// donate [#mobId|@userId] !itemId
				if parts[0] == `detonate` {

					if len(parts) == 1 {
						continue
					}

					var itemName string
					var targetName string

					if len(parts) == 2 {
						itemName = parts[1]
					} else if len(parts) == 3 {
						targetName = parts[1]
						itemName = parts[2]
					}

					if itm, found := room.FindOnFloor(itemName, false); found {

						iSpec := itm.GetSpec()
						if iSpec.Type == items.Grenade {

							room.RemoveItem(itm, false)

							messageQueue.SendRoomMessage(actionRequest.RoomId, fmt.Sprintf(`The <ansi fg="itemname">%s</ansi> <ansi fg="red">EXPLODES</ansi>!`, itm.Name()), true)

							hitMobs := true
							hitPlayers := true

							targetPlayerId, targetMobId := room.FindByName(targetName)

							if targetPlayerId > 0 {
								hitMobs = false
							}

							if targetMobId > 0 {
								hitPlayers = false
							}

							if hitPlayers {
								for _, uid := range room.GetPlayers() {
									for _, buffId := range iSpec.BuffIds {
										w.QueueBuff(uid, 0, buffId)
									}
								}
							}

							if hitMobs {
								for _, mid := range room.GetMobs() {

									for _, buffId := range iSpec.BuffIds {
										w.QueueBuff(0, mid, buffId)
									}

									if actionRequest.SourceUserId > 0 {

										if sourceUser := users.GetByUserId(actionRequest.SourceUserId); sourceUser != nil {

											if mob := mobs.GetInstance(mid); mob != nil {
												mob.DamageTaken[sourceUser.UserId] = 0 // Take note that the player did damage this mob.

												if sourceUser.Character.RoomId == mob.Character.RoomId {
													// Mobs get aggro when attacked
													if mob.Character.Aggro == nil {
														mob.PreventIdle = true
														w.QueueCommand(0, mid, fmt.Sprintf("attack @%d", sourceUser.UserId)) // @ means player
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
																w.QueueCommand(0, mid, fmt.Sprintf("go %s", exitName))
																w.QueueCommand(0, mid, fmt.Sprintf("attack @%d", sourceUser.UserId)) // @ means player
																break
															}
														}
													}

													if foundExitName != `` {
														w.QueueCommand(0, mid, fmt.Sprintf("go %s", foundExitName))
														w.QueueCommand(0, mid, fmt.Sprintf("attack @%d", sourceUser.UserId)) // @ means player
													}
												}
											}

										}

									}

								}
							}

						}

					}
				}
			}

		}
	}

	// Process any pending buffs on mobs
	for w.mobBuffQueue.Len(0) > 0 {
		if buffRequest, ok := w.mobBuffQueue.Pop(); ok {
			// Apply the buff
			if buffInfo := buffs.GetBuffSpec(buffRequest.BuffId); buffInfo != nil {

				if buffMob := mobs.GetInstance(buffRequest.ToId); buffMob != nil {

					buffMob.Character.AddBuff(buffRequest.BuffId)

					if len(buffInfo.Messages.Start.Room) > 0 {
						roomMsg := buffInfo.Messages.Start.Room
						roomMsg = strings.ReplaceAll(roomMsg, `{username}'s`, fmt.Sprintf(`<ansi fg="mobname">%s's</ansi>`, buffMob.Character.Name))
						roomMsg = strings.ReplaceAll(roomMsg, `{username}`, fmt.Sprintf(`<ansi fg="mobname">%s</ansi>`, buffMob.Character.Name))
						messageQueue.SendRoomMessage(buffMob.Character.RoomId, roomMsg, true)
					}

					if buffInfo.TriggerNow {
						msgs := buffMob.Character.TriggerBuffs(buffInfo.BuffId)

						for _, msgInfo := range msgs.Messages {

							msgInfo.Msg = strings.ReplaceAll(msgInfo.Msg, `{username}'s`, fmt.Sprintf(`<ansi fg="mobname">%s's</ansi>`, buffMob.Character.Name))
							msgInfo.Msg = strings.ReplaceAll(msgInfo.Msg, `{username}`, fmt.Sprintf(`<ansi fg="mobname">%s</ansi>`, buffMob.Character.Name))

							if msgInfo.ToRoom {
								messageQueue.SendRoomMessage(buffMob.Character.RoomId, msgInfo.Msg, false)
							}
						}

						if buffMob.Character.Health <= 0 {
							// Mob died
							w.QueueCommand(0, buffMob.InstanceId, `suicide`)
						}
					}

				}
			}
		}
	}

	// Process any pending buffs on users
	for w.userBuffQueue.Len(0) > 0 {
		if buffRequest, ok := w.userBuffQueue.Pop(); ok {
			// Apply the buff

			if buffInfo := buffs.GetBuffSpec(buffRequest.BuffId); buffInfo != nil {

				if buffUser := users.GetByUserId(buffRequest.ToId); buffUser != nil {

					//
					// Buff removal
					//
					if buffRequest.BuffId < 0 {

						buffUser.Character.RemoveBuff(buffInfo.BuffId * -1)

					} else {

						//
						// Add buff
						//
						buffUser.Character.AddBuff(buffRequest.BuffId)

						if len(buffInfo.Messages.Start.User) > 0 {
							messageQueue.SendUserMessage(buffUser.UserId, buffInfo.Messages.Start.User, true)
						}

						if len(buffInfo.Messages.Start.Room) > 0 {
							roomMsg := buffInfo.Messages.Start.Room
							roomMsg = strings.ReplaceAll(roomMsg, "{username}'s", fmt.Sprintf(`<ansi fg="username">%s's</ansi>`, buffUser.Character.Name))
							roomMsg = strings.ReplaceAll(roomMsg, "{username}", fmt.Sprintf(`<ansi fg="username">%s</ansi>`, buffUser.Character.Name))
							messageQueue.SendRoomMessage(buffUser.Character.RoomId, roomMsg, true, buffUser.UserId)
						}

						if buffInfo.TriggerNow {
							msgs := buffUser.Character.TriggerBuffs(buffInfo.BuffId)

							for _, msgInfo := range msgs.Messages {

								msgInfo.Msg = strings.ReplaceAll(msgInfo.Msg, `{username}'s`, fmt.Sprintf(`<ansi fg="username">%s's</ansi>`, buffUser.Character.Name))
								msgInfo.Msg = strings.ReplaceAll(msgInfo.Msg, `{username}`, fmt.Sprintf(`<ansi fg="username">%s</ansi>`, buffUser.Character.Name))

								if msgInfo.ToRoom {
									messageQueue.SendRoomMessage(buffUser.Character.RoomId, msgInfo.Msg, false, buffUser.UserId)
								} else {
									messageQueue.SendUserMessage(buffUser.UserId, msgInfo.Msg, false)
								}
							}
						}
					}

				}
			}
		}
	}

	// handle queued quest toke handouts
	for w.userQuestQueue.Len(0) > 0 {
		if questRequest, ok := w.userQuestQueue.Pop(); ok {

			// Give them a token
			remove := false
			if questRequest.QuestToken[0:1] == `-` {
				remove = true
				questRequest.QuestToken = questRequest.QuestToken[1:]
			}

			if questInfo := quests.GetQuest(questRequest.QuestToken); questInfo != nil {

				if questUser := users.GetByUserId(questRequest.ToId); questUser != nil {

					if remove {
						questUser.Character.ClearQuestToken(questRequest.QuestToken)
						continue
					}

					// This only succees if the user doesn't have the quest yet or the quest is a later step of one they've started
					if questUser.Character.GiveQuestToken(questRequest.QuestToken) {

						_, stepName := quests.TokenToParts(questRequest.QuestToken)

						if stepName == `start` {
							if !questInfo.Secret {
								questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You have been given a new quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name))
								messageQueue.SendUserMessage(questUser.UserId, questUpTxt, true)
							}
						} else if stepName == `end` {

							if !questInfo.Secret {
								questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You have completed the quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name))
								messageQueue.SendUserMessage(questUser.UserId, questUpTxt, true)
							}

							// Message to player?
							if len(questInfo.Rewards.PlayerMessage) > 0 {
								messageQueue.SendUserMessage(questUser.UserId, questInfo.Rewards.PlayerMessage, true)
							}
							// Message to room?
							if len(questInfo.Rewards.RoomMessage) > 0 {
								messageQueue.SendRoomMessage(questUser.Character.RoomId, questInfo.Rewards.RoomMessage, true, questUser.UserId)
							}
							// New quest to start?
							if len(questInfo.Rewards.QuestId) > 0 {
								w.QueueQuest(questUser.UserId, questInfo.Rewards.QuestId)
							}
							// Gold reward?
							if questInfo.Rewards.Gold > 0 {
								messageQueue.SendUserMessage(questUser.UserId, fmt.Sprintf(`You receive <ansi fg="gold">%d gold</ansi>!`, questInfo.Rewards.Gold), true)
								questUser.Character.Gold += questInfo.Rewards.Gold
							}
							// Item reward?
							if questInfo.Rewards.ItemId > 0 {
								newItm := items.New(questInfo.Rewards.ItemId)
								messageQueue.SendUserMessage(questUser.UserId, fmt.Sprintf(`You receive <ansi fg="itemname">%s</ansi>!`, newItm.NameSimple()), true)
								questUser.Character.StoreItem(newItm)

								iSpec := newItm.GetSpec()
								if iSpec.QuestToken != `` {
									w.QueueQuest(questUser.UserId, iSpec.QuestToken)
								}
							}
							// Buff reward?
							if questInfo.Rewards.BuffId > 0 {
								w.QueueBuff(questUser.UserId, 0, questInfo.Rewards.BuffId)
							}
							// Experience reward?
							if questInfo.Rewards.Experience > 0 {

								grantXP, xpScale := questUser.Character.GrantXP(questInfo.Rewards.Experience)

								xpMsgExtra := ``
								if xpScale != 100 {
									xpMsgExtra = fmt.Sprintf(` <ansi fg="yellow">(%d%% scale)</ansi>`, xpScale)
								}

								messageQueue.SendUserMessage(questUser.UserId, fmt.Sprintf(`You receive <ansi fg="experience">%d experience points</ansi>%s!`, grantXP, xpMsgExtra), true)
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
										messageQueue.SendUserMessage(questUser.UserId, skillUpTxt, true)
									}

								}
							}
							// Move them to another room/area?
							if questInfo.Rewards.RoomId > 0 {
								messageQueue.SendUserMessage(questUser.UserId, `You are suddenly moved to a new place!`, true)
								messageQueue.SendRoomMessage(questUser.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> is suddenly moved to a new place!`, questUser.Character.Name), true, questUser.UserId)
								rooms.MoveToRoom(questUser.UserId, questInfo.Rewards.RoomId)
							}
						} else {
							if !questInfo.Secret {
								questUpTxt, _ := templates.Process("character/questup", fmt.Sprintf(`You've made progress on the quest: <ansi fg="questname">%s</ansi>!`, questInfo.Name))
								messageQueue.SendUserMessage(questUser.UserId, questUpTxt, true)
							}
						}

					}

				}

			}
		}
	}

	//
	// Prune all buffs that have expired.
	//
	messageQueue.AbsorbMessages(w.PruneBuffs())

	if turnCt%uint64(c.TurnsPerSecond()) == 0 {
		messageQueue.AbsorbMessages(w.CheckForLevelUps())
	}

	if messageQueue.Pending() {
		w.DispatchMessages(messageQueue)
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

func NewWorld(osSignalChan chan os.Signal) *World {

	w := &World{
		connectionPool: connection.New(osSignalChan),
		users:          users.NewUserManager(),
		worldInput:     make(chan WorldInput),
		userInputQueue: util.LimitQueue[WorldInput]{},
		mobInputQueue:  util.LimitQueue[WorldInput]{},
		userBuffQueue:  util.LimitQueue[BuffApply]{},
		mobBuffQueue:   util.LimitQueue[BuffApply]{},
	}

	w.userInputQueue.SetLimit(10)

	return w
}
