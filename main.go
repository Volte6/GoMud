package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"log/slog"

	"github.com/Volte6/ansitags"
	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/gametime"
	"github.com/volte6/mud/inputhandlers"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/quests"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/term"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
	"github.com/volte6/mud/version"
	"github.com/volte6/mud/webclient"
)

const (
	// Version is the current version of the server
	Version = `1.0.0`
)

var (
	logger = slog.New(
		util.GetColorLogHandler(os.Stderr, slog.LevelDebug),
	)

	sigChan            = make(chan os.Signal, 1)
	workerShutdownChan = make(chan bool, 1)

	serverAlive atomic.Bool

	worldManager = NewWorld(sigChan)

	// Start a pool of worker goroutines
	wg sync.WaitGroup
)

func main() {

	// Setup the default logger
	slog.SetDefault(logger)

	configs.ReloadConfig()
	c := configs.GetConfig()

	slog.Info(`========================`)
	//
	slog.Info(`  ___  ____   _______   `)
	slog.Info(`  |  \/  | | | |  _  \  `)
	slog.Info(`  | .  . | | | | | | |  `)
	slog.Info(`  | |\/| | | | | | | |  `)
	slog.Info(`  | |  | | |_| | |/ /   `)
	slog.Info(`  \_|  |_/\___/|___/    `)
	//
	slog.Info(`========================`)
	//
	cfgData := c.AllConfigData()
	cfgKeys := make([]string, 0, len(cfgData))
	for k := range cfgData {
		cfgKeys = append(cfgKeys, k)
	}

	// sort the keys
	slices.Sort(cfgKeys)

	for _, k := range cfgKeys {
		slog.Info("Config", k, cfgData[k])
	}
	//
	slog.Info(`========================`)

	// Do version related checks
	slog.Info(`Version: ` + Version)
	if err := version.VersionCheck(Version); err != nil {

		if err == version.ErrIncompatibleVersion {
			slog.Error("Incompatible version.", "details", "Backup all datafiles and run with -u or --upgrade flag to attempt an automatic upgrade.")
			return
		}

		if err == version.ErrUpgradePossible {
			slog.Warn("Version mismatch.", "details", "Your config files could use some updating. Backup all datafiles and run with -u or --upgrade flag to attempt an automatic upgrade.")
		}

	}
	slog.Info(`========================`)

	//
	// System Configurations
	runtime.GOMAXPROCS(int(c.MaxCPUCores))

	// Load all the data files up front.
	spells.LoadSpellFiles()
	rooms.LoadDataFiles()
	buffs.LoadDataFiles() // Load buffs before items for cost calculation reasons
	items.LoadDataFiles()
	races.LoadDataFiles()
	mobs.LoadDataFiles()
	quests.LoadDataFiles()
	templates.LoadAliases()
	keywords.LoadAliases()
	gametime.SetToDay(-5)

	for name, str := range items.GetDebugColorPatternOutput() {
		slog.Info("Color Pattern Test", name, ansitags.Parse(str))
	}

	scripting.Setup(int(c.ScriptLoadTimeoutMs), int(c.ScriptRoomTimeoutMs))

	//
	slog.Info(`========================`)
	//
	// Capture OS signals to gracefully shutdown the server
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//
	// Spin up server listeners
	//

	// Set the server to be alive
	serverAlive.Store(true)

	webclient.Listen(int(c.WebPort), &wg)

	allTelnetPorts := strings.Split(string(c.TelnetPort), `,`)

	allServerListeners := make([]net.Listener, 0, len(allTelnetPorts))
	for _, port := range allTelnetPorts {
		if p, err := strconv.Atoi(port); err == nil {
			if s := TelnetListenOnPort(``, p, &wg, int(c.MaxTelnetConnections)); s != nil {
				allServerListeners = append(allServerListeners, s)
			}
		}
	}

	if c.LocalPort > 0 {
		TelnetListenOnPort(`127.0.0.1`, int(c.LocalPort), &wg, 0)
	}

	go worldManager.InputWorker(workerShutdownChan, &wg)
	go worldManager.MaintenanceWorker(workerShutdownChan, &wg)
	go worldManager.GameTickWorker(workerShutdownChan, &wg)

	// block until a signal comes in
	<-sigChan

	tplTxt, err := templates.Process("goodbye", nil)
	if err != nil {
		slog.Error("Template Error", "error", err)
	}
	worldManager.GetConnectionPool().Broadcast([]byte(tplTxt))

	serverAlive.Store(false) // immediately stop processing incoming connections

	// some last minute stats reporting
	totalConnections, totalDisconnections := worldManager.GetConnectionPool().Stats()
	slog.Error(
		"shutting down server",
		"LifetimeConnections", totalConnections,
		"LifetimeDisconnects", totalDisconnections,
		"ActiveConnections", totalConnections-totalDisconnections,
	)

	// cleanup all connections
	worldManager.GetConnectionPool().Cleanup()

	for _, s := range allServerListeners {
		s.Close()
	}

	webclient.Shutdown()

	// Just an ephemeral goroutine that spins its wheels until the program shuts down")
	go func() {
		for {
			slog.Error("Waiting on workers")
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

}

func handleTelnetConnection(connDetails *connection.ConnectionDetails, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	slog.Info("New Connection", "connectionID", connDetails.ConnectionId(), "remoteAddr", connDetails.RemoteAddr().String())

	// Add starting handlers

	// Special escape handlers
	connDetails.AddInputHandler("TelnetIACHandler", inputhandlers.TelnetIACHandler)
	connDetails.AddInputHandler("AnsiHandler", inputhandlers.AnsiHandler)
	// Consider a macro handler at this point?
	// Text Processing
	connDetails.AddInputHandler("CleanserInputHandler", inputhandlers.CleanserInputHandler)
	connDetails.AddInputHandler("LoginInputHandler", inputhandlers.LoginInputHandler)

	// Turn off "line at a time", send chars as typed
	worldManager.GetConnectionPool().SendTo(
		term.TelnetWILL(term.TELNET_OPT_SUP_GO_AHD),
		connDetails.ConnectionId(),
	)
	// Tell the client we expect chars as they are typed
	worldManager.GetConnectionPool().SendTo(
		term.TelnetWONT(term.TELNET_OPT_LINE_MODE),
		connDetails.ConnectionId(),
	)

	// Tell the client we intend to echo back what they type
	// So they shouldn't locally echo it

	worldManager.GetConnectionPool().SendTo(
		term.TelnetWILL(term.TELNET_OPT_ECHO),
		connDetails.ConnectionId(),
	)
	// Request that the client report window size changes as they happen
	worldManager.GetConnectionPool().SendTo(
		term.TelnetDO(term.TELNET_OPT_NAWS),
		connDetails.ConnectionId(),
	)

	// Can separate with a space multiple charsets:
	// "UTF-8 ISO-8859-1"
	worldManager.GetConnectionPool().SendTo(
		term.TelnetCharset.BytesWithPayload([]byte(" UTF-8")),
		connDetails.ConnectionId(),
	)

	clientSetupCommands := "" + //term.AnsiAltModeStart.String() + // alternative mode (No scrollback)
		//term.AnsiCursorHide.String() + // Hide Cursor (Because we will manually echo back)
		//term.AnsiCharSetUTF8.String() + // UTF8 mode
		//term.AnsiReportMouseClick.String() + // Request client to capture and report mouse clicks
		term.AnsiRequestResolution.String() // Request resolution
		//""

	worldManager.GetConnectionPool().SendTo(
		[]byte(clientSetupCommands),
		connDetails.ConnectionId(),
	)

	// an input buffer for reading data sent over the network
	inputBuffer := make([]byte, connection.ReadBufferSize)

	// Describes whatever the client sent us
	clientInput := &connection.ClientInput{
		ConnectionId: connDetails.ConnectionId(),
		DataIn:       []byte{},
		Buffer:       make([]byte, 0, connection.ReadBufferSize), // DataIn is appended to this buffer after processing
		EnterPressed: false,
		Clipboard:    []byte{},
		History:      connection.InputHistory{},
	}

	var sharedState map[string]any = make(map[string]any)

	// Invoke the login handler for the first time
	// The default behavior is to just send a welcome screen first
	inputhandlers.LoginInputHandler(clientInput, worldManager.GetConnectionPool(), sharedState)

	var userObject *users.UserRecord
	var suggestions Suggestions
	lastInput := time.Now()
	for {

		c := configs.GetConfig()

		clientInput.EnterPressed = false // Default state is always false
		clientInput.TabPressed = false   // Default state is always false
		clientInput.BSPressed = false    // Default state is always false

		n, err := connDetails.Read(inputBuffer)
		if err != nil {

			// If failed to read from the connection, switch to zombie state
			if userObject != nil {

				if c.ZombieSeconds > 0 {

					connDetails.SetState(connection.Zombie)
					users.SetZombieUser(userObject.UserId)

				} else {

					worldManager.LeaveWorld(userObject.UserId)

					if err := users.LogOutUserByConnectionId(connDetails.ConnectionId()); err != nil {
						slog.Error("Log Out Error", "connectionId", connDetails.ConnectionId(), "error", err)
					}

				}
			}

			if err == io.EOF {
				worldManager.GetConnectionPool().Remove(connDetails.ConnectionId())
			} else {
				slog.Warn("Conn Read Error", "error", err)
			}

			break
		}

		if connDetails.InputDisabled() {
			continue
		}

		clientInput.DataIn = inputBuffer[:n]

		// Input handler processes any special commands, transforms input, sets flags from input, etc
		okContinue, lastHandler, err := connDetails.HandleInput(clientInput, worldManager.GetConnectionPool(), sharedState)

		// Was there an error? If so, we should probably just stop processing input
		if err != nil {
			logger.Warn("InputHandler", "error", err)
			continue
		}

		// If a handler aborted processing, just keep track of where we are so
		// far and jump back to waiting.
		if !okContinue {
			if userObject != nil {

				_, suggested := userObject.GetUnsentText()

				redrawPrompt := false

				if clientInput.TabPressed {

					if suggestions.Count() < 1 {
						suggestions.Set(worldManager.GetAutoComplete(userObject.UserId, string(clientInput.Buffer)))
					}

					if suggestions.Count() > 0 {
						suggested = suggestions.Next()
						userObject.SetUnsentText(string(clientInput.Buffer), suggested)
						redrawPrompt = true
					}

				} else if clientInput.BSPressed {
					// If a suggestion is pending, remove it
					// otherwise just do a normal backspace operation
					userObject.SetUnsentText(string(clientInput.Buffer), ``)
					if suggested != `` {
						suggested = ``
						suggestions.Clear()
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
							suggestions.Clear()
						} else {
							suggested = ``
							suggestions.Clear()
							// Otherwise, just keep the suggestion
							userObject.SetUnsentText(string(clientInput.Buffer), suggested)
							redrawPrompt = true
						}
					}

					userObject.SetUnsentText(string(clientInput.Buffer), suggested)
				}

				if redrawPrompt {
					worldManager.GetConnectionPool().SendTo([]byte(templates.AnsiParse(userObject.GetCommandPrompt(true))), clientInput.ConnectionId)
				}

			}
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

			if userObject.Permission == users.PermissionAdmin {
				connDetails.AddInputHandler("AdminCommandInputHandler", inputhandlers.AdminCommandInputHandler)
			}

			connDetails.AddInputHandler("SystemCommandInputHandler", inputhandlers.SystemCommandInputHandler)

			// Add a signal handler (shortcut ctrl combos) after the AnsiHandler
			// This captures signals and replaces user input so should happen after AnsiHandler to ensure it happens before other processes.
			connDetails.AddInputHandler("SignalHandler", inputhandlers.SignalHandler, "AnsiHandler")

			connDetails.SetState(connection.LoggedIn)

			worldManager.EnterWorld(userObject.Character.RoomId, userObject.Character.Zone, userObject.UserId)
		}

		// If they have pressed enter (submitted their input), and nothing else has handled/aborted
		if clientInput.EnterPressed {

			if time.Since(lastInput) < time.Duration(c.TurnMs)*time.Millisecond {
				/*
					worldManager.GetConnectionPool().SendTo(
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
					suggestions.Clear()
					userObject.SetUnsentText(string(clientInput.Buffer), ``)
					worldManager.GetConnectionPool().SendTo([]byte(templates.AnsiParse(userObject.GetCommandPrompt(true))), clientInput.ConnectionId)

				}

				wi := WorldInput{
					FromId:    userObject.UserId,
					InputText: string(clientInput.Buffer),
				}

				// Buffer should be processed as an in-game command
				worldManager.Input(wi)
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

func TelnetListenOnPort(hostname string, portNum int, wg *sync.WaitGroup, maxConnections int) net.Listener {

	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostname, portNum))
	if err != nil {
		slog.Error("Error creating server", "error", err)
		return nil
	}

	// Start a goroutine to accept incoming connections, so that we can use a signal to stop the server
	go func() {

		// Loop to accept connections
		for {
			conn, err := server.Accept()

			if !serverAlive.Load() {
				slog.Error("Connections disabled.")
				return
			}

			if err != nil {
				slog.Error("Connection error", "error", err)
				continue
			}

			if maxConnections > 0 {
				if worldManager.GetConnectionPool().ActiveConnectionCount() >= maxConnections {
					conn.Write([]byte(fmt.Sprintf("\n\n\n!!! Server is full (%d connections). Try again later. !!!\n\n\n", worldManager.GetConnectionPool().ActiveConnectionCount())))
					conn.Close()
					continue
				}
			}

			wg.Add(1)
			// hand off the connection to a handler goroutine so that we can continue handling new connections
			go handleTelnetConnection(
				worldManager.GetConnectionPool().Add(conn),
				wg,
			)

		}
	}()

	return server
}
