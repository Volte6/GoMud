package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/volte6/gomud/internal/badinputtracker"
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/keywords"
	"github.com/volte6/gomud/internal/mobcommands"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/prompt"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/usercommands"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
	"github.com/volte6/gomud/internal/web"
)

type WorldInput struct {
	FromId    int
	InputText string
	ReadyTurn uint64
}

func (wi WorldInput) Id() int {
	return wi.FromId
}

type World struct {
	worldInput         chan WorldInput
	ignoreInput        map[int]uint64 // userid->turn set to ignore
	enterWorldUserId   chan [2]int
	leaveWorldUserId   chan int
	logoutConnectionId chan connections.ConnectionId
	zombieFlag         chan [2]int
}

func NewWorld(osSignalChan chan os.Signal) *World {

	w := &World{
		worldInput:         make(chan WorldInput),
		ignoreInput:        make(map[int]uint64),
		enterWorldUserId:   make(chan [2]int),
		leaveWorldUserId:   make(chan int),
		logoutConnectionId: make(chan connections.ConnectionId),
		zombieFlag:         make(chan [2]int),
	}

	connections.SetShutdownChan(osSignalChan)

	return w
}

// Send input to the world.
// Just sends via a channel. Will block until read.
func (w *World) SendInput(i WorldInput) {
	w.worldInput <- i
}

func (w *World) SendEnterWorld(userId int, roomId int) {
	w.enterWorldUserId <- [2]int{userId, roomId}
}

func (w *World) SendLeaveWorld(userId int) {
	w.leaveWorldUserId <- userId
}

func (w *World) SendLogoutConnectionId(connId connections.ConnectionId) {
	w.logoutConnectionId <- connId
}

func (w *World) SendSetZombie(userId int, on bool) {
	if on {
		w.zombieFlag <- [2]int{userId, 1}
	} else {
		w.zombieFlag <- [2]int{userId, 0}
	}
}

func (w *World) logOutUserByConnectionId(connectionId connections.ConnectionId) {

	if err := users.LogOutUserByConnectionId(connectionId); err != nil {
		mudlog.Error("Log Out Error", "connectionId", connectionId, "error", err)
	}
}

func (w *World) enterWorld(userId int, roomId int) {

	if userInfo := users.GetByUserId(userId); userInfo != nil {
		events.AddToQueue(events.PlayerSpawn{
			UserId:        userInfo.UserId,
			RoomId:        userInfo.Character.RoomId,
			Username:      userInfo.Username,
			CharacterName: userInfo.Character.Name,
		})
	}

	w.UpdateStats()

	// Put htme in the room
	rooms.MoveToRoom(userId, roomId, true)
}

/*
users can be:
Disconnected	+ OutWorld (no presence)	No record in connections.netConnections or users.ZombieConnections	| user object in room
Connected		+ OutWorld (logging in) 	Has record in connections.netConnections 							| user object in room
Connected		+ InWorld  (non-zombie) 	No record in users.ZombieConnections								| no zombie flag		| user object in room
Disconnected	+ InWorld  (zombie)			Has record in users.ZombieConnections 								| has zombie flag		| user object in room
*/

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

func (w *World) MainWorker(shutdown chan bool, wg *sync.WaitGroup) {

	wg.Add(1)

	mudlog.Info("MainWorker", "state", "Started")
	defer func() {
		mudlog.Warn("MainWorker", "state", "Stopped")
		wg.Done()
	}()

	c := configs.GetConfig()

	roomUpdateTimer := time.NewTimer(roomMaintenancePeriod)
	ansiAliasTimer := time.NewTimer(ansiAliasReloadPeriod)
	eventLoopTimer := time.NewTimer(time.Millisecond)
	turnTimer := time.NewTimer(time.Duration(c.TurnMs) * time.Millisecond)
	statsTimer := time.NewTimer(time.Duration(10) * time.Second)

loop:
	for {

		// The reason for
		// util.LockGame() / util.UnlockGame()
		// In each of these cases is to lock down the
		// logic for when other processes need to query data
		// such as the webserver

		select {
		case <-shutdown:

			mudlog.Warn(`MainWorker`, `action`, `shutdown received`)

			util.LockMud()
			if err := rooms.SaveAllRooms(); err != nil {
				mudlog.Error("rooms.SaveAllRooms()", "error", err.Error())
			}
			users.SaveAllUsers() // Save all user data too.
			util.UnlockMud()

			break loop
		case <-statsTimer.C:

			// TODO: Move this to events
			util.LockMud()

			w.UpdateStats()
			// save the round counter.
			util.SaveRoundCount(c.FolderDataFiles.String() + `/` + util.RoundCountFilename)

			util.UnlockMud()

			statsTimer.Reset(time.Duration(10) * time.Second)

		case <-roomUpdateTimer.C:
			mudlog.Debug(`MainWorker`, `action`, `rooms.RoomMaintenance()`)

			// TODO: Move this to events
			util.LockMud()
			rooms.RoomMaintenance()
			util.UnlockMud()

			roomUpdateTimer.Reset(roomMaintenancePeriod)

		case <-ansiAliasTimer.C:

			// TODO: Move this to events
			util.LockMud()
			templates.LoadAliases()
			util.UnlockMud()

			ansiAliasTimer.Reset(ansiAliasReloadPeriod)

		case <-eventLoopTimer.C:

			eventLoopTimer.Reset(time.Millisecond)

			util.LockMud()
			w.EventLoop()
			util.UnlockMud()

		case <-turnTimer.C:

			util.LockMud()
			turnTimer.Reset(time.Duration(c.TurnMs) * time.Millisecond)

			turnCt := util.IncrementTurnCount()
			c := configs.GetConfig()

			events.AddToQueue(events.NewTurn{TurnNumber: turnCt, TimeNow: time.Now(), Config: c})

			// After a full round of turns, we can do a round tick.
			if turnCt%uint64(c.TurnsPerRound()) == 0 {

				roundNumber := util.IncrementRoundCount()

				if c.LogIntervalRoundCount > 0 && roundNumber%uint64(c.LogIntervalRoundCount) == 0 {
					mudlog.Info("World::RoundTick()", "roundNumber", roundNumber)
				}

				events.AddToQueue(events.NewRound{RoundNumber: roundNumber, TimeNow: time.Now(), Config: c})
			}

			w.EventLoopTurns()

			util.UnlockMud()

		case enterWorldUserId := <-w.enterWorldUserId: // [2]int

			util.LockMud()
			w.enterWorld(enterWorldUserId[0], enterWorldUserId[1])
			util.UnlockMud()

		case leaveWorldUserId := <-w.leaveWorldUserId: // int

			util.LockMud()
			if userInfo := users.GetByUserId(leaveWorldUserId); userInfo != nil {
				events.AddToQueue(events.PlayerDespawn{
					UserId:        userInfo.UserId,
					RoomId:        userInfo.Character.RoomId,
					Username:      userInfo.Username,
					CharacterName: userInfo.Character.Name,
				})
			}
			util.UnlockMud()

		case logoutConnectionId := <-w.logoutConnectionId: //  connections.ConnectionId

			util.LockMud()
			w.logOutUserByConnectionId(logoutConnectionId)
			util.UnlockMud()

		case zombieFlag := <-w.zombieFlag: //  [2]int
			if zombieFlag[1] == 1 {

				util.LockMud()
				users.SetZombieUser(zombieFlag[0])
				util.UnlockMud()

			}
		}
		c = configs.GetConfig()
	}

}

// Should be goroutine/threadsafe
// Only reads from world channel
func (w *World) InputWorker(shutdown chan bool, wg *sync.WaitGroup) {
	wg.Add(1)

	mudlog.Info("InputWorker", "state", "Started")
	defer func() {
		mudlog.Warn("InputWorker", "state", "Stopped")
		wg.Done()
	}()

loop:
	for {
		select {
		case <-shutdown:
			mudlog.Warn(`InputWorker`, `action`, `shutdown received`)
			break loop
		case wi := <-w.worldInput:

			events.AddToQueue(events.Input{
				UserId:    wi.FromId,
				InputText: wi.InputText,
				ReadyTurn: util.GetTurnCount(),
			})

		}
	}
}

func (w *World) processInput(userId int, inputText string, flags events.EventFlag) {

	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		mudlog.Error("User not found", "userId", userId)
		return
	}

	var activeQuestion *prompt.Question = nil
	hadPrompt := false
	if cmdPrompt := user.GetPrompt(); cmdPrompt != nil {
		hadPrompt = true
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
				readyTurn := util.GetTurnCount()
				for _, newCmd := range strings.Split(macro, `;`) {
					if newCmd == `` {
						continue
					}

					events.AddToQueue(events.Input{
						UserId:    userId,
						InputText: newCmd,
						ReadyTurn: readyTurn,
					})

					readyTurn++
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

			handled, err = usercommands.TryCommand(command, remains, userId, flags)
			if err != nil {
				mudlog.Warn("user-TryCommand", "command", command, "remains", remains, "error", err.Error())
			}
		}

	} else {
		connId := user.ConnectionId()
		connections.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt(true))), connId)
	}

	if !handled {
		if len(command) > 0 {

			badinputtracker.TrackBadCommand(command, remains)

			user.SendText(fmt.Sprintf(`<ansi fg="command">%s</ansi> not recognized. Type <ansi fg="command">help</ansi> for commands.`, command))
			user.Command(`emote @looks a little confused`)
		}
	}

	// If they had an input prompt, but now they don't, lets make sure to resend a status prompt
	if hadPrompt || (!hadPrompt && user.GetPrompt() != nil) {
		connId := user.ConnectionId()
		connections.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt(true))), connId)
	}
	// Removing this as possibly redundant.
	// Leaving in case I need to remember that I did it...
	//connId := user.ConnectionId()
	//connections.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt(true))), connId)

}

func (w *World) processMobInput(mobInstanceId int, inputText string) {
	// No need to select the channel this way

	mob := mobs.GetInstance(mobInstanceId)
	if mob == nil { // Something went wrong. User not found.
		if !mobs.RecentlyDied(mobInstanceId) {
			mudlog.Error("Mob not found", "mobId", mobInstanceId, "where", "processMobInput()")
		}
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

		//mudlog.Info("World received mob input", "InputText", (inputText))

		handled, err = mobcommands.TryCommand(command, remains, mobInstanceId)
		if err != nil {
			mudlog.Warn("mob-TryCommand", "command", command, "remains", remains, "error", err.Error())
		}

	}

	if !handled {
		if len(command) > 0 {
			mob.Command(fmt.Sprintf(`emote looks a little confused (%s %s).`, command, remains))
		}
	}

}

// Events that are throttled by TurnMs in config.yaml
func (w *World) EventLoopTurns() {

	var turnCt uint64 = util.GetTurnCount()

	//
	// Handle Input Queue
	//
	alreadyProcessed := make(map[int]struct{}) // Keep track of players who already had a command this turn, and what turn it was
	eq := events.GetQueue(events.Input{})
	for eq.Len() > 0 {

		e := eq.Poll()

		input, typeOk := e.(events.Input)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "Input", "Actual Type", e.Type())
			continue
		}

		//mudlog.Debug(`Event`, `type`, input.Type(), `UserId`, input.UserId, `MobInstanceId`, input.MobInstanceId, `WaitTurns`, input.WaitTurns, `InputText`, input.InputText)

		// If it's a mob
		if input.MobInstanceId > 0 {
			if input.ReadyTurn <= turnCt {

				// Allow any handlers to handle the event
				if !events.DoListeners(e) {
					continue
				}

				w.processMobInput(input.MobInstanceId, input.InputText)
			} else {
				events.Requeue(input)
			}
			continue
		}

		// 0 and below, process immediately and don't count towards limit
		if input.ReadyTurn == 0 {

			// If this command was potentially blocking input, unblock it now.
			if input.Flags.Has(events.CmdUnBlockInput) {

				if _, ok := w.ignoreInput[input.UserId]; ok {
					delete(w.ignoreInput, input.UserId)
					if user := users.GetByUserId(input.UserId); user != nil {
						user.UnblockInput()
					}
				}

			}

			// Allow any handlers to handle the event
			if !events.DoListeners(e) {
				continue
			}

			w.processInput(input.UserId, input.InputText, input.Flags)

			continue
		}

		// If an event was already processed for this user this turn, skip
		if _, ok := alreadyProcessed[input.UserId]; ok {
			events.Requeue(input)
			continue
		}

		// 0 means process immediately
		// however, process no further events from this user until next turn
		if input.ReadyTurn > turnCt {

			// If this is a multi-turn wait, block further input if flagged to do so
			if input.Flags.Has(events.CmdBlockInput) {

				if _, ok := w.ignoreInput[input.UserId]; !ok {
					w.ignoreInput[input.UserId] = turnCt
				}

				input.Flags.Remove(events.CmdBlockInput)
			}

			events.Requeue(input)

			continue
		}

		//
		// Event ready to be processed
		//

		// If this command was potentially blocking input, unblock it now.
		if input.Flags.Has(events.CmdUnBlockInput) {

			if _, ok := w.ignoreInput[input.UserId]; ok {
				delete(w.ignoreInput, input.UserId)
				if user := users.GetByUserId(input.UserId); user != nil {
					user.UnblockInput()
				}
			}

		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		w.processInput(input.UserId, input.InputText, events.EventFlag(input.Flags))

		alreadyProcessed[input.UserId] = struct{}{}
	}

}

// Should only handle sending messages out to users
func (w *World) EventLoop() {

	var c configs.Config = configs.GetConfig()

	turnNow := util.GetTurnCount()

	// Player joined the world
	//
	eq := events.GetQueue(events.Log{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Auctions
	//
	eq = events.GetQueue(events.Auction{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Day/Night handling
	//
	eq = events.GetQueue(events.DayNightCycle{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Player joined the world
	//
	eq = events.GetQueue(events.PlayerSpawn{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Player left the world
	//
	eq = events.GetQueue(events.PlayerDespawn{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Player Levelup Notifications
	//
	eq = events.GetQueue(events.LevelUp{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Death Notifications
	//
	eq = events.GetQueue(events.PlayerDeath{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// ScriptedEvents
	//
	eq = events.GetQueue(events.ScriptedEvent{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// ItemOwnership
	//
	eq = events.GetQueue(events.ItemOwnership{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// System commands such as /reload
	//
	eq = events.GetQueue(events.System{})
	for eq.Len() > 0 {

		e := eq.Poll()

		sys, typeOk := e.(events.System)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "System", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		if sys.Command == `reload` {

			events.AddToQueue(events.Broadcast{
				Text: `Reloading flat files...`,
			})

			loadAllDataFiles(true)

			events.AddToQueue(events.Broadcast{
				Text:            `Done.` + term.CRLFStr,
				SkipLineRefresh: true,
			})

		} else if sys.Command == `kick` {
			w.Kick(sys.Data.(int))
		} else if sys.Command == `leaveworld` {

			if userInfo := users.GetByUserId(sys.Data.(int)); userInfo != nil {
				events.AddToQueue(events.PlayerDespawn{
					UserId:        userInfo.UserId,
					RoomId:        userInfo.Character.RoomId,
					Username:      userInfo.Username,
					CharacterName: userInfo.Character.Name,
				})
			}

		} else if sys.Command == `logoff` {
			w.logOff(sys.Data.(int))
		}

	}

	//
	// Handle Turn Queue
	//
	eq = events.GetQueue(events.NewTurn{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Handle RoomAction Queue
	// Needs a major overhaul/change to how it works.
	//
	eq = events.GetQueue(events.RoomAction{})
	for eq.Len() > 0 {

		e := eq.Poll()

		action, typeOk := e.(events.RoomAction)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "RoomAction", "Actual Type", e.Type())
			continue
		}

		//mudlog.Debug(`Event`, `type`, action.Type(), `RoomId`, action.RoomId, `SourceUserId`, action.SourceUserId, `SourceMobId`, action.SourceMobId, `WaitTurns`, action.WaitTurns, `Action`, action.Action)

		if action.ReadyTurn > turnNow {

			if int(action.ReadyTurn-turnNow)%c.TurnsPerRound() == 0 {
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

			events.Requeue(action)
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		// Make sure the room exists
		room := rooms.LoadRoom(action.RoomId)
		if room == nil {
			continue
		}

		if action.Action == `mutator` {

			mutName := action.Details.(string)

			if mutName == `wildfire` {
				if room.GetBiome().Burns() && room.Mutators.Add(mutName) {
					room.SendText(colorpatterns.ApplyColorPattern(`A wildfire burns through the area!`, `flame`, colorpatterns.Stretch))
				}
				continue
			}

			room.Mutators.Add(mutName)

			continue
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
				RoomId:  room.RoomId,
				Action:  `mutator`,
				Details: `Details`,
			})

			if hitPlayers {

				for _, uid := range room.GetPlayers() {

					// If not hitting self and pvp is disabled, skip
					if action.SourceUserId > 0 && action.SourceUserId != uid && configs.GetConfig().PVP != `enabled` {
						continue
					}

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

				mob.Character.TrackPlayerDamage(sourceUser.UserId, 0)

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
		events.DoListeners(eq.Poll())
	}

	//
	// Handle Quest Queue
	//
	eq = events.GetQueue(events.Quest{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Handle NewRound events
	//
	eq = events.GetQueue(events.NewRound{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Dispatch GMCP events
	//
	eq = events.GetQueue(events.GMCPOut{})
	for eq.Len() > 0 {

		e := eq.Poll()

		gmcp, typeOk := e.(events.GMCPOut)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "GMCPOut", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		if gmcp.UserId < 1 {
			continue
		}

		connId := users.GetConnectionId(gmcp.UserId)
		if connId == 0 {
			continue
		}

		switch v := gmcp.Payload.(type) {
		case []byte:
			connections.SendTo(term.GmcpPayload.BytesWithPayload(v), connId)
		case string:
			connections.SendTo(term.GmcpPayload.BytesWithPayload([]byte(v)), connId)
		default:
			payload, err := json.Marshal(gmcp.Payload)
			if err != nil {
				mudlog.Error("Event", "Type", "GMCPOut", "data", gmcp.Payload, "error", err)
				continue
			}
			connections.SendTo(term.GmcpPayload.BytesWithPayload(payload), connId)
		}

	}

	//
	// Handle RoomChange events
	//
	eq = events.GetQueue(events.RoomChange{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	// Dispatch MSP events
	//
	eq = events.GetQueue(events.MSP{})
	for eq.Len() > 0 {
		events.DoListeners(eq.Poll())
	}

	//
	//
	// What follows are communication based events
	// (Events expected to send output to users)
	//
	//
	redrawPrompts := make(map[uint64]string)

	//
	// System-wide broadcasts
	//
	eq = events.GetQueue(events.Broadcast{})
	for eq.Len() > 0 {

		e := eq.Poll()

		broadcast, typeOk := e.(events.Broadcast)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "Broadcast", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		messageColorized := templates.AnsiParse(broadcast.Text)

		var sentToConnectionIds []connections.ConnectionId

		//
		// If it's communication, respect deafeaning rules
		//
		skipConnectionIds := []connections.ConnectionId{}
		if broadcast.IsCommunication {
			for _, u := range users.GetAllActiveUsers() {
				if u.Deafened && !broadcast.SourceIsMod {
					skipConnectionIds = append(skipConnectionIds, u.ConnectionId())
				}
			}
		}

		if broadcast.SkipLineRefresh {

			sentToConnectionIds = connections.Broadcast(
				[]byte(messageColorized),
				skipConnectionIds...,
			)

		} else {

			sentToConnectionIds = connections.Broadcast(
				[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+messageColorized),
				skipConnectionIds...,
			)

		}

		for _, connId := range sentToConnectionIds {
			if _, ok := redrawPrompts[connId]; !ok {
				user := users.GetByConnectionId(connId)
				redrawPrompts[connId] = templates.AnsiParse(user.GetCommandPrompt(true))
			}
		}
	}

	eq = events.GetQueue(events.WebClientCommand{})
	for eq.Len() > 0 {

		e := eq.Poll()

		cmd, typeOk := e.(events.WebClientCommand)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		if !connections.IsWebsocket(cmd.ConnectionId) {
			continue
		}

		connections.SendTo([]byte(cmd.Text), cmd.ConnectionId)

	}

	//
	// Outbound text strings
	//
	eq = events.GetQueue(events.Message{})
	for eq.Len() > 0 {

		e := eq.Poll()

		message, typeOk := e.(events.Message)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
			continue
		}

		// Allow any handlers to handle the event
		if !events.DoListeners(e) {
			continue
		}

		//mudlog.Debug("Message{}", "userId", message.UserId, "roomId", message.RoomId, "length", len(messageColorized), "IsCommunication", message.IsCommunication)

		if message.UserId > 0 {

			if user := users.GetByUserId(message.UserId); user != nil {

				// If they are deafened, they cannot hear user communications
				if message.IsCommunication && user.Deafened {
					continue
				}

				connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+templates.AnsiParse(message.Text)), user.ConnectionId())
				if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
					redrawPrompts[user.ConnectionId()] = templates.AnsiParse(user.GetCommandPrompt(true))
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

					// If they are deafened, they cannot hear user communications
					if message.IsCommunication && user.Deafened {
						continue
					}

					// If this is a quiet message, make sure the player can hear it
					if message.IsQuiet {
						if !user.Character.HasBuffFlag(buffs.SuperHearing) {
							continue
						}
					}

					connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+templates.AnsiParse(message.Text)), user.ConnectionId())
					if _, ok := redrawPrompts[user.ConnectionId()]; !ok {
						redrawPrompts[user.ConnectionId()] = templates.AnsiParse(user.GetCommandPrompt(true))
					}

				}
			}

		}

	}

	for connectionId, prompt := range redrawPrompts {
		connections.SendTo([]byte(prompt), connectionId)
	}
}

func (w *World) UpdateStats() {
	s := web.GetStats()
	s.Reset()

	c := configs.GetConfig()

	for _, u := range users.GetAllActiveUsers() {
		s.OnlineUsers = append(s.OnlineUsers, u.GetOnlineInfo())
	}

	sort.Slice(s.OnlineUsers, func(i, j int) bool {
		if s.OnlineUsers[i].Permission == users.PermissionAdmin {
			return true
		}
		if s.OnlineUsers[j].Permission == users.PermissionAdmin {
			return false
		}
		return s.OnlineUsers[i].OnlineTime > s.OnlineUsers[j].OnlineTime
	})

	for _, t := range c.TelnetPort {
		p, _ := strconv.Atoi(t)
		if p > 0 {
			s.TelnetPorts = append(s.TelnetPorts, p)
		}
	}

	s.WebSocketPort = int(c.WebPort)

	web.UpdateStats(s)
}

// Force disconnect a user (Makes them a zombie)
func (w *World) Kick(userId int) {

	mudlog.Info(`Kick`, `userId`, userId)

	user := users.GetByUserId(userId)
	if user == nil {
		return
	}
	users.SetZombieUser(userId)

	user.EventLog.Add(`conn`, `Kicked`)

	connections.Kick(user.ConnectionId())
}
