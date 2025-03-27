package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/volte6/gomud/internal/audio"
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/flags"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/hooks"
	"github.com/volte6/gomud/internal/inputhandlers"
	"github.com/volte6/gomud/internal/integrations/discord"
	"github.com/volte6/gomud/internal/items"
	"github.com/volte6/gomud/internal/keywords"
	"github.com/volte6/gomud/internal/language"

	"github.com/volte6/gomud/internal/mapper"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/mutators"
	"github.com/volte6/gomud/internal/pets"
	"github.com/volte6/gomud/internal/plugins"
	"github.com/volte6/gomud/internal/quests"
	"github.com/volte6/gomud/internal/races"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/spells"
	"github.com/volte6/gomud/internal/suggestions"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
	"github.com/volte6/gomud/internal/util"
	"github.com/volte6/gomud/internal/web"
	_ "github.com/volte6/gomud/modules"
	textLang "golang.org/x/text/language"
)

const (
	// Version is the current version of the server
	Version = `1.0.0`
)

var (
	sigChan            = make(chan os.Signal, 1)
	workerShutdownChan = make(chan bool, 1)

	serverAlive atomic.Bool

	worldManager = NewWorld(sigChan)

	// Start a pool of worker goroutines
	wg sync.WaitGroup
)

func main() {

	// Capture panic and write msg/stack to logs
	defer func() {
		if r := recover(); r != nil {
			mudlog.Error("PANIC", "error", r)
			s := string(debug.Stack())
			for _, str := range strings.Split(s, "\n") {
				mudlog.Error("PANIC", "stack", str)
			}
		}
	}()

	// Setup logging
	mudlog.SetupLogger(
		events.GetLogger(),
		os.Getenv(`LOG_LEVEL`),
		os.Getenv(`LOG_PATH`),
		os.Getenv(`LOG_NOCOLOR`) == ``,
	)

	flags.HandleFlags()

	configs.ReloadConfig()
	c := configs.GetConfig()

	// Default i18n localize folders
	if len(c.Translation.LanguagePaths) == 0 {
		c.Translation.LanguagePaths = []string{
			path.Join("_datafiles", "localize"),
			path.Join(c.FilePaths.DataFiles.String(), "localize"),
		}
	}

	mudlog.Info(`========================`)
	//
	mudlog.Info(`  ___  ____   _______   `)
	mudlog.Info(`  |  \/  | | | |  _  \  `)
	mudlog.Info(`  | .  . | | | | | | |  `)
	mudlog.Info(`  | |\/| | | | | | | |  `)
	mudlog.Info(`  | |  | | |_| | |/ /   `)
	mudlog.Info(`  \_|  |_/\___/|___/    `)
	//
	mudlog.Info(`========================`)
	//
	cfgData := c.AllConfigData()
	cfgKeys := make([]string, 0, len(cfgData))
	for k := range cfgData {
		cfgKeys = append(cfgKeys, k)
	}

	// sort the keys
	slices.Sort(cfgKeys)
	for _, k := range cfgKeys {
		mudlog.Info("Config", "name", k, "value", cfgData[k])
	}
	//
	mudlog.Info(`========================`)

	// Register the plugin filesystem with the template system
	templates.RegisterFS(plugins.GetPluginRegistry())

	//
	// System Configurations
	runtime.GOMAXPROCS(int(c.Server.MaxCPUCores))

	// Validate chosen world:
	if err := util.ValidateWorldFiles(`_datafiles/world/default`, c.FilePaths.DataFiles.String()); err != nil {
		mudlog.Error("World Validation", "error", err)
		os.Exit(1)
	}

	language.InitTranslation(language.BundleCfg{
		DefaultLanguage: textLang.Make(c.Translation.DefaultLanguage.String()),
		Language:        textLang.Make(c.Translation.Language.String()),
		LanguagePaths:   c.Translation.LanguagePaths,
	})

	hooks.RegisterListeners()

	// Discord integration
	if webhookUrl := string(c.Integrations.Discord.WebhookUrl); webhookUrl != "" {
		discord.Init(webhookUrl)
		mudlog.Info("Discord", "info", "integration is enabled")
	} else {
		mudlog.Warn("Discord", "info", "integration is disabled")
	}

	mudlog.Error(
		"Starting server",
		"name", string(c.Server.MudName),
	)

	// Load all the data files up front.
	loadAllDataFiles(false)

	mudlog.Info("Mapper", "status", "precaching")
	timeStart := time.Now()
	mapper.PreCacheMaps()
	mudlog.Info("Mapper", "status", "done", "time taken", time.Since(timeStart))

	// Create the user index
	idx := users.NewUserIndex()
	if !idx.Exists() {
		// Since it doesn't exist yet, that's a good indication we should do a quick format migration check
		users.DoUserMigrations()
	}
	idx.Create()
	idx.Rebuild()
	mudlog.Info("UserIndex", "info", "User index recreated.")

	// Load the round count from the file
	if util.LoadRoundCount(c.FilePaths.DataFiles.String()+`/`+util.RoundCountFilename) == util.RoundCountMinimum {
		gametime.SetToDay(-3)
	}

	gametime.GetZodiac(1) // The first time this is called it randomizes all zodiacs

	scripting.Setup(int(c.Scripting.LoadTimeoutMs), int(c.Scripting.RoomTimeoutMs))

	//
	mudlog.Info(`========================`)

	// Trigger the load plugins event
	plugins.Load(
		configs.GetFilePathsConfig().DataFiles.String(),
	)

	web.SetWebPlugin(plugins.GetPluginRegistry())

	//
	// Capture OS signals to gracefully shutdown the server
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//
	// Spin up server listeners
	//

	// Set the server to be alive
	serverAlive.Store(true)

	web.Listen(int(c.Network.WebPort), int(c.Network.HttpsPort), &wg, HandleWebSocketConnection)

	allServerListeners := make([]net.Listener, 0, len(c.Network.TelnetPort))
	for _, port := range c.Network.TelnetPort {
		if p, err := strconv.Atoi(port); err == nil {
			if s := TelnetListenOnPort(``, p, &wg, int(c.Network.MaxTelnetConnections)); s != nil {
				allServerListeners = append(allServerListeners, s)
			}
		}
	}

	if c.Network.LocalPort > 0 {
		TelnetListenOnPort(`127.0.0.1`, int(c.Network.LocalPort), &wg, 0)
	}

	go worldManager.InputWorker(workerShutdownChan, &wg)
	go worldManager.MainWorker(workerShutdownChan, &wg)
	//go worldManager.MaintenanceWorker(workerShutdownChan, &wg)
	//go worldManager.GameTickWorker(workerShutdownChan, &wg)

	// block until a signal comes in
	<-sigChan

	tplTxt, err := templates.Process("goodbye", nil, templates.AnsiTagsPreParse)
	if err != nil {
		mudlog.Error("Template Error", "error", err)
	}

	events.AddToQueue(events.Broadcast{
		Text: templates.AnsiParse(tplTxt),
	})

	serverAlive.Store(false) // immediately stop processing incoming connections

	util.SaveRoundCount(c.FilePaths.DataFiles.String() + `/` + util.RoundCountFilename)

	// some last minute stats reporting
	totalConnections, totalDisconnections := connections.Stats()
	mudlog.Error(
		"Stopping server",
		"LifetimeConnections", totalConnections,
		"LifetimeDisconnects", totalDisconnections,
		"ActiveConnections", totalConnections-totalDisconnections,
	)

	// cleanup all connections
	connections.Cleanup()

	for _, s := range allServerListeners {
		s.Close()
	}

	web.Shutdown()

	// Final plugin save before shutting down
	plugins.Save()

	// Just an ephemeral goroutine that spins its wheels until the program shuts down")
	go func() {
		for {
			mudlog.Warn("Waiting on workers")
			// sleep for 3 seconds
			time.Sleep(time.Duration(3) * time.Second)
		}
	}()

	// Send the worker shutdown signal for each worker thread to read
	workerShutdownChan <- true
	workerShutdownChan <- true
	workerShutdownChan <- true

	// Wait for all workers to finish their tasks.
	// Otherwise we end up getting flushed file saves incomplete.
	wg.Wait()

	// Give it a second to disaptch any final messages in the event queue
	// Example: discord server shutdown
	time.Sleep(1 * time.Second)
}

func handleTelnetConnection(connDetails *connections.ConnectionDetails, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	mudlog.Info("New Connection", "connectionID", connDetails.ConnectionId(), "remoteAddr", connDetails.RemoteAddr().String())

	// Add starting handlers

	// Special escape handlers
	connDetails.AddInputHandler("TelnetIACHandler", inputhandlers.TelnetIACHandler)
	connDetails.AddInputHandler("AnsiHandler", inputhandlers.AnsiHandler)
	// Consider a macro handler at this point?
	// Text Processing
	connDetails.AddInputHandler("CleanserInputHandler", inputhandlers.CleanserInputHandler)
	connDetails.AddInputHandler("LoginInputHandler", inputhandlers.LoginInputHandler)

	// Turn off "line at a time", send chars as typed
	connections.SendTo(
		term.TelnetWILL(term.TELNET_OPT_SUP_GO_AHD),
		connDetails.ConnectionId(),
	)
	// Tell the client we expect chars as they are typed
	connections.SendTo(
		term.TelnetWONT(term.TELNET_OPT_LINE_MODE),
		connDetails.ConnectionId(),
	)

	// Tell the client we intend to echo back what they type
	// So they shouldn't locally echo it

	connections.SendTo(
		term.TelnetWILL(term.TELNET_OPT_ECHO),
		connDetails.ConnectionId(),
	)
	// Request that the client report window size changes as they happen
	connections.SendTo(
		term.TelnetDO(term.TELNET_OPT_NAWS),
		connDetails.ConnectionId(),
	)

	// Send request to change charset
	connections.SendTo(
		term.TelnetRequestChangeCharset.BytesWithPayload(nil),
		connDetails.ConnectionId(),
	)

	// Send request to enable GMCP
	connections.SendTo(
		term.GmcpEnable.BytesWithPayload(nil),
		connDetails.ConnectionId(),
	)

	// Send request to enable MSP
	connections.SendTo(
		term.MspEnable.BytesWithPayload(nil),
		connDetails.ConnectionId(),
	)

	connections.SendTo(
		term.TelnetSuppressGoAhead.BytesWithPayload(nil),
		connDetails.ConnectionId(),
	)

	clientSetupCommands := "" + //term.AnsiAltModeStart.String() + // alternative mode (No scrollback)
		//term.AnsiCursorHide.String() + // Hide Cursor (Because we will manually echo back)
		//term.AnsiCharSetUTF8.String() + // UTF8 mode
		//term.AnsiReportMouseClick.String() + // Request client to capture and report mouse clicks
		term.AnsiRequestResolution.String() // Request resolution
		//""

	connections.SendTo(
		[]byte(clientSetupCommands),
		connDetails.ConnectionId(),
	)

	// an input buffer for reading data sent over the network
	inputBuffer := make([]byte, connections.ReadBufferSize)

	// Describes whatever the client sent us
	clientInput := &connections.ClientInput{
		ConnectionId: connDetails.ConnectionId(),
		DataIn:       []byte{},
		Buffer:       make([]byte, 0, connections.ReadBufferSize), // DataIn is appended to this buffer after processing
		EnterPressed: false,
		Clipboard:    []byte{},
		History:      connections.InputHistory{},
	}

	var sharedState map[string]any = make(map[string]any)

	// Invoke the login handler for the first time
	// The default behavior is to just send a welcome screen first
	inputhandlers.LoginInputHandler(clientInput, sharedState)

	if audioConfig := audio.GetFile(`intro`); audioConfig.FilePath != `` {
		v := 100
		if audioConfig.Volume > 0 && audioConfig.Volume <= 100 {
			v = audioConfig.Volume
		}
		connections.SendTo(
			term.MspCommand.BytesWithPayload([]byte("!!MUSIC("+audioConfig.FilePath+" V="+strconv.Itoa(v)+" L=-1 C=1)")),
			clientInput.ConnectionId,
		)
	}

	var userObject *users.UserRecord
	var sug suggestions.Suggestions
	lastInput := time.Now()
	c := configs.GetConfig()

	for {

		clientInput.EnterPressed = false // Default state is always false
		clientInput.TabPressed = false   // Default state is always false
		clientInput.BSPressed = false    // Default state is always false

		n, err := connDetails.Read(inputBuffer)
		if err != nil {

			// If failed to read from the connection, switch to zombie state
			if userObject != nil {

				userObject.EventLog.Add(`conn`, `Disconnected`)

				if c.Network.ZombieSeconds > 0 {

					connDetails.SetState(connections.Zombie)
					worldManager.SendSetZombie(userObject.UserId, true)

				} else {

					worldManager.SendLeaveWorld(userObject.UserId)
					worldManager.SendLogoutConnectionId(connDetails.ConnectionId())

				}

			}

			mudlog.Warn("Telnet", "error", err)

			connections.Remove(connDetails.ConnectionId())

			break
		}

		if connDetails.InputDisabled() {
			continue
		}

		clientInput.DataIn = inputBuffer[:n]

		// Input handler processes any special commands, transforms input, sets flags from input, etc
		okContinue, lastHandler, err := connDetails.HandleInput(clientInput, sharedState)

		// Was there an error? If so, we should probably just stop processing input
		if err != nil {
			mudlog.Warn("InputHandler", "error", err)
			continue
		}

		// If a handler aborted processing, just keep track of where we are so
		// far and jump back to waiting.
		if !okContinue {
			if userObject != nil {

				_, suggested := userObject.GetUnsentText()

				redrawPrompt := false

				if clientInput.TabPressed {

					if sug.Count() < 1 {
						sug.Set(worldManager.GetAutoComplete(userObject.UserId, string(clientInput.Buffer)))
					}

					if sug.Count() > 0 {
						suggested = sug.Next()
						userObject.SetUnsentText(string(clientInput.Buffer), suggested)
						redrawPrompt = true
					}

				} else if clientInput.BSPressed {
					// If a suggestion is pending, remove it
					// otherwise just do a normal backspace operation
					userObject.SetUnsentText(string(clientInput.Buffer), ``)
					if suggested != `` {
						suggested = ``
						sug.Clear()
						redrawPrompt = true
					}

				} else {

					if suggested != `` {

						// If they hit space, accept the suggestion
						if len(clientInput.Buffer) > 0 && clientInput.Buffer[len(clientInput.Buffer)-1] == term.ASCII_SPACE {
							clientInput.Buffer = append(clientInput.Buffer[0:len(clientInput.Buffer)-1], []byte(suggested)...)
							clientInput.Buffer = append(clientInput.Buffer[0:len(clientInput.Buffer)], []byte(` `)...)
							redrawPrompt = true
							userObject.SetUnsentText(string(clientInput.Buffer), ``)
							sug.Clear()
						} else {
							suggested = ``
							sug.Clear()
							// Otherwise, just keep the suggestion
							userObject.SetUnsentText(string(clientInput.Buffer), suggested)
							redrawPrompt = true
						}
					}

					userObject.SetUnsentText(string(clientInput.Buffer), suggested)
				}

				if redrawPrompt {
					pTxt := userObject.GetCommandPrompt()
					if connections.IsWebsocket(clientInput.ConnectionId) {
						connections.SendTo([]byte(pTxt), clientInput.ConnectionId)
					} else {
						connections.SendTo([]byte(templates.AnsiParse(pTxt)), clientInput.ConnectionId)
					}
				}

			}
			continue
		}

		if lastHandler == "LoginInputHandler" {

			connections.SendTo(
				term.MspCommand.BytesWithPayload([]byte("!!MUSIC(Off)")),
				clientInput.ConnectionId,
			)

			// Remove the login handler
			connDetails.RemoveInputHandler("LoginInputHandler")
			// Replace it with a regular echo handler.
			connDetails.AddInputHandler("EchoInputHandler", inputhandlers.EchoInputHandler)
			// Add admin command handler
			connDetails.AddInputHandler("HistoryInputHandler", inputhandlers.HistoryInputHandler) // Put history tracking after login handling, since login handling aborts input until complete

			if val, ok := sharedState["LoginInputHandler"]; ok {
				state := val.(*inputhandlers.LoginState)
				userObject = state.UserObject
			}

			if userObject.Role == users.RoleAdmin {
				connDetails.AddInputHandler("SystemCommandInputHandler", inputhandlers.SystemCommandInputHandler)
			}

			// Add a signal handler (shortcut ctrl combos) after the AnsiHandler
			// This captures signals and replaces user input so should happen after AnsiHandler to ensure it happens before other processes.
			connDetails.AddInputHandler("SignalHandler", inputhandlers.SignalHandler, "AnsiHandler")

			connDetails.SetState(connections.LoggedIn)

			worldManager.SendEnterWorld(userObject.UserId, userObject.Character.RoomId)

		}

		// If they have pressed enter (submitted their input), and nothing else has handled/aborted
		if clientInput.EnterPressed {

			// Update config after enter presses
			// No need to update it every loop
			c = configs.GetConfig()

			if time.Since(lastInput) < time.Duration(c.Timing.TurnMs)*time.Millisecond {
				/*
					connections.SendTo(
						[]byte("Slow down! You're typing too fast! "+time.Since(lastInput).String()+"\n"),
						connDetails.ConnectionId(),
					)
				*/

				// Reset the buffer for future commands.
				clientInput.Reset()

				// Capturing and resetting the unsent text is purely to allow us to
				// Keep updating the prompt without losing the typed in text.
				userObject.SetUnsentText(``, ``)

			} else {

				_, suggested := userObject.GetUnsentText()

				if len(suggested) > 0 {
					// solidify it in the render for UX reasons

					clientInput.Buffer = append(clientInput.Buffer, []byte(suggested)...)
					sug.Clear()
					userObject.SetUnsentText(string(clientInput.Buffer), ``)

					if connections.IsWebsocket(clientInput.ConnectionId) {
						connections.SendTo([]byte(userObject.GetCommandPrompt()), clientInput.ConnectionId)
					} else {
						connections.SendTo([]byte(templates.AnsiParse(userObject.GetCommandPrompt())), clientInput.ConnectionId)
					}

				}

				wi := WorldInput{
					FromId:    userObject.UserId,
					InputText: string(clientInput.Buffer),
				}

				// Buffer should be processed as an in-game command
				worldManager.SendInput(wi)
				// Reset the buffer for future commands.
				clientInput.Reset()

				// Capturing and resetting the unsent text is purely to allow us to
				// Keep updating the prompt without losing the typed in text.
				userObject.SetUnsentText(``, ``)

				lastInput = time.Now()
			}

			time.Sleep(time.Duration(10) * time.Millisecond)
			//	time.Sleep(time.Duration(util.TurnMs) * time.Millisecond)
		}

	}

}

func HandleWebSocketConnection(conn *websocket.Conn) {

	var userObject *users.UserRecord
	connDetails := connections.Add(nil, conn)
	connDetails.AddInputHandler("LoginInputHandler", inputhandlers.LoginInputHandler)

	// Describes whatever the client sent us
	clientInput := &connections.ClientInput{
		ConnectionId: connDetails.ConnectionId(),
		DataIn:       []byte{},
		Buffer:       make([]byte, 0, connections.ReadBufferSize), // DataIn is appended to this buffer after processing
		EnterPressed: false,
		Clipboard:    []byte{},
		History:      connections.InputHistory{},
	}

	var sharedState map[string]any = make(map[string]any)

	// Invoke the login handler for the first time
	// The default behavior is to just send a welcome screen first
	inputhandlers.LoginInputHandler(clientInput, sharedState)

	connections.SendTo(
		[]byte("!!SOUND(Off U="+configs.GetConfig().FilePaths.WebCDNLocation.String()+")"),
		clientInput.ConnectionId,
	)

	if audioConfig := audio.GetFile(`intro`); audioConfig.FilePath != `` {
		v := 100
		if audioConfig.Volume > 0 && audioConfig.Volume <= 100 {
			v = audioConfig.Volume
		}
		connections.SendTo(
			[]byte("!!MUSIC("+audioConfig.FilePath+" V="+strconv.Itoa(v)+" L=-1 C=1)"),
			clientInput.ConnectionId,
		)
	}

	c := configs.GetConfig()

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {

			// If failed to read from the connection, switch to zombie state
			if userObject != nil {

				userObject.EventLog.Add(`conn`, `Disconnected`)

				if c.Network.ZombieSeconds > 0 {

					connDetails.SetState(connections.Zombie)
					worldManager.SendSetZombie(userObject.UserId, true)

				} else {

					worldManager.SendLeaveWorld(userObject.UserId)
					worldManager.SendLogoutConnectionId(connDetails.ConnectionId())

				}

			}

			mudlog.Warn("WS Read", "error", err)
			break
		}

		clientInput.DataIn = message
		clientInput.Buffer = message
		clientInput.EnterPressed = true

		// Input handler processes any special commands, transforms input, sets flags from input, etc
		okContinue, lastHandler, err := connDetails.HandleInput(clientInput, sharedState)
		if !okContinue {
			continue
		}

		if lastHandler == "LoginInputHandler" {
			// Remove the login handler
			connDetails.RemoveInputHandler("LoginInputHandler")
			// Replace it with a regular echo handler.
			connDetails.AddInputHandler("EchoInputHandler", inputhandlers.EchoInputHandler)
			// Add admin command handler
			connDetails.AddInputHandler("HistoryInputHandler", inputhandlers.HistoryInputHandler) // Put history tracking after login handling, since login handling aborts input until complete

			if val, ok := sharedState["LoginInputHandler"]; ok {
				state := val.(*inputhandlers.LoginState)
				userObject = state.UserObject
			}

			if userObject.Role == users.RoleAdmin {
				connDetails.AddInputHandler("SystemCommandInputHandler", inputhandlers.SystemCommandInputHandler)
			}

			// Add a signal handler (shortcut ctrl combos) after the AnsiHandler
			// This captures signals and replaces user input so should happen after AnsiHandler to ensure it happens before other processes.
			connDetails.AddInputHandler("SignalHandler", inputhandlers.SignalHandler, "AnsiHandler")

			connDetails.SetState(connections.LoggedIn)

			worldManager.SendEnterWorld(userObject.UserId, userObject.Character.RoomId)

			continue
		}

		wi := WorldInput{
			FromId:    userObject.UserId,
			InputText: string(message),
		}

		// Buffer should be processed as an in-game command
		worldManager.SendInput(wi)

		c = configs.GetConfig()
	}
}

func TelnetListenOnPort(hostname string, portNum int, wg *sync.WaitGroup, maxConnections int) net.Listener {

	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostname, portNum))
	if err != nil {
		mudlog.Error("Error creating server", "error", err)
		return nil
	}

	// Start a goroutine to accept incoming connections, so that we can use a signal to stop the server
	go func() {

		// Loop to accept connections
		for {
			conn, err := server.Accept()

			if !serverAlive.Load() {
				mudlog.Warn("Connections disabled.")
				return
			}

			if err != nil {
				mudlog.Warn("Connection error", "error", err)
				continue
			}

			if maxConnections > 0 {
				if connections.ActiveConnectionCount() >= maxConnections {
					conn.Write([]byte(fmt.Sprintf("\n\n\n!!! Server is full (%d connections). Try again later. !!!\n\n\n", connections.ActiveConnectionCount())))
					conn.Close()
					continue
				}
			}

			wg.Add(1)
			// hand off the connection to a handler goroutine so that we can continue handling new connections
			go handleTelnetConnection(
				connections.Add(conn, nil),
				wg,
			)

		}
	}()

	return server
}

func loadAllDataFiles(isReload bool) {

	if isReload {

		defer func() {
			if r := recover(); r != nil {
				mudlog.Error("RELOAD FAILED", "err", r)
			}
		}()

	}

	// Force clear all cached VM's
	scripting.PruneVMs(true)

	spells.LoadSpellFiles()
	rooms.LoadDataFiles()
	buffs.LoadDataFiles() // Load buffs before items for cost calculation reasons
	items.LoadDataFiles()
	races.LoadDataFiles()
	mobs.LoadDataFiles()
	pets.LoadDataFiles()
	quests.LoadDataFiles()
	templates.LoadAliases()
	keywords.LoadAliases(plugins.GetPluginRegistry())
	mutators.LoadDataFiles()
	colorpatterns.LoadColorPatterns()
	audio.LoadAudioConfig()
	characters.CompileAdjectiveSwaps() // This should come after loading color patterns.
}
