package inputhandlers

import (
	"fmt"
	"strings"

	"log/slog"

	"github.com/volte6/gomud/connections"
	"github.com/volte6/gomud/templates"
	"github.com/volte6/gomud/users"
)

func SystemCommandInputHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	// if they didn't hit enter, just keep buffering, go next.
	if !clientInput.EnterPressed {
		return true
	}

	// Handle text input

	// grab whatever was typed in so far and trim out leading/trailing whitespace and null byte
	message := strings.TrimSpace(string(clientInput.Buffer))

	//slog.Info("SystemCommandInputHandler Received", "type", "TXT", "size", len(clientInput.Buffer), "data", string(clientInput.Buffer), "message", message)

	// If all they ever sent was white space, we won't have anything to work with. We can just ignore the input...
	// ALternatively, they may have only hit ENTER, and we may do something with that...
	if len(message) == 0 {
		return true
	}

	// If successful, we're done.
	if trySystemCommand(message, clientInput.ConnectionId) {
		// zero out the current buffer
		clientInput.Buffer = clientInput.Buffer[:0]
		return false
	}

	return true
}

// TODO: Move into own handler?

const SystemCommandPrefix = "/"

type SystemCommandHelp struct {
	Description  string
	Details      string
	ExampleInput string
}

var (
	systemCommandList = map[string]SystemCommandHelp{
		"quit": SystemCommandHelp{
			Description:  "Disconnect self from the server",
			ExampleInput: "quit",
		},
		"who": SystemCommandHelp{
			Description:  "List all connected users",
			ExampleInput: "who",
		},
		"help": SystemCommandHelp{
			Description:  "Display this help message, or help for a specific command",
			ExampleInput: "help shutdown",
		},
	}
)

func systemCommandParts(cmd string) (systemCmd string, cmdArg string) {

	cmd = strings.TrimSpace(cmd)

	// If there's a space split it into cmd vs arg
	if index := strings.Index(cmd, " "); index != -1 {

		systemCmd, cmdArg = strings.ToLower(cmd[0:index]), cmd[index+1:]

		if cmdArg[0:1] == " " {
			cmdArg = strings.TrimSpace(cmdArg)
		}

		return systemCmd, cmdArg
	}

	systemCmd, cmdArg = strings.ToLower(cmd), ""

	return systemCmd, cmdArg
}

func trySystemCommand(cmd string, connectionId connections.ConnectionId) bool {

	if len(cmd) < 1 {
		return false
	}

	if cmd[0:1] != SystemCommandPrefix {
		return false
	}

	cmd, arg := systemCommandParts(strings.TrimSpace(cmd[1:]))

	// look for cmd in the command list
	if _, ok := systemCommandList[cmd]; !ok {
		return false
	}

	slog.Info("system command", "cmd", cmd, "arg", arg)
	//fmt.Printf("cmd:[%s] arg:[%s]\n", cmd, arg)

	if cmd == "quit" {
		// Not building complex output, so just preparse the ansi in the template and cache that
		tplTxt, _ := templates.Process("goodbye", nil, templates.AnsiTagsPreParse)

		if connections.IsWebsocket(connectionId) {
			connections.SendTo([]byte(tplTxt), connectionId)
		} else {
			connections.SendTo([]byte(templates.AnsiParse(tplTxt)), connectionId)
		}

		connections.Kick(connectionId)
		return true
	}

	if cmd == "who" {
		onlineUsers := users.GetOnlineList()

		headers := []string{"UserId", "Username", "Character", "Level", "Role"}

		rows := [][]string{}
		for _, user := range onlineUsers {
			rows = append(rows, []string{user.UserId, user.Username, user.CharacterName, fmt.Sprintf(`%d`, user.CharacterLevel), user.Permission})
		}

		onlineTableData := templates.GetTable("Online Users", headers, rows)
		tplTxt, _ := templates.Process("tables/generic", onlineTableData)

		if connections.IsWebsocket(connectionId) {
			connections.SendTo([]byte(tplTxt), connectionId)
		} else {
			connections.SendTo([]byte(templates.AnsiParse(tplTxt)), connectionId)
		}

		// Not building complex output, so just preparse the ansi in the template and cache that
		//tplTxt, _ := templates.Process("systemcommands/who", onlineUsers)
		//connections.SendTo([]byte(tplTxt), connectionId)
		return true
	}

	if cmd == "help" {

		tplTxt, _ := templates.Process("systemcommands/help", systemCommandList)

		if connections.IsWebsocket(connectionId) {
			connections.SendTo([]byte(tplTxt), connectionId)
		} else {
			connections.SendTo([]byte(templates.AnsiParse(tplTxt)), connectionId)
		}

		return true
	}

	slog.Error("valid command unhandled", "cmd", cmd, "arg", arg)

	return true
}
