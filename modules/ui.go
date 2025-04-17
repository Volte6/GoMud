package modules

import (
	"embed"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var (
	//go:embed ui/*
	ui_Files embed.FS
	
	// Instance accessible from elsewhere in this package
	uiModule UIModule
)

func init() {
	// Create the module instance
	uiModule = UIModule{
		plug: plugins.New(`ui`, `1.0`),
	}

	// Attach the file system to the plugin
	if err := uiModule.plug.AttachFileSystem(ui_Files); err != nil {
		panic(err)
	}
	
	// Set default values
	uiModule.setDefaultValues()
	
	// Register the onLoad callback to load config values after plugin system is ready
	uiModule.plug.Callbacks.SetOnLoad(func() {
		uiModule.RefreshConfig()
	})

	// Register the user command
	uiModule.plug.AddUserCommand(`gui`, GUICommand, true, false)
}

// UIModule stores configuration and state for the UI module
type UIModule struct {
	plug        *plugins.Plugin
	uiUrl       string
	uiVersion   string
	removeField string // Field name for removal flag
}

// Set hardcoded default values for initialization
func (g *UIModule) setDefaultValues() {
	g.uiUrl = "https://github.com/MorquinDevlar/GoMudUI/releases/latest/download/gomud-ui.mpackage"
	g.uiVersion = "1.0.0"
	g.removeField = "remove"
}

// RefreshConfig loads values from the module's config.yaml file
func (g *UIModule) RefreshConfig() {
	// Using direct config properties like other modules do
	if uiUrl, ok := g.plug.Config.Get(`UIUrl`).(string); ok && uiUrl != "" {
		g.uiUrl = uiUrl
		mudlog.Info("UIModule", "config", "Loaded UIUrl from config", "value", g.uiUrl)
	} else {
		mudlog.Info("UIModule", "config", "Using default UIUrl", "value", g.uiUrl)
	}

	if uiVersion, ok := g.plug.Config.Get(`UIVersion`).(string); ok && uiVersion != "" {
		g.uiVersion = uiVersion
		mudlog.Info("UIModule", "config", "Loaded UIVersion from config", "value", g.uiVersion)
	} else {
		mudlog.Info("UIModule", "config", "Using default UIVersion", "value", g.uiVersion)
	}
	
	if removeField, ok := g.plug.Config.Get(`RemoveField`).(string); ok && removeField != "" {
		g.removeField = removeField
		mudlog.Info("UIModule", "config", "Loaded RemoveField from config", "value", g.removeField)
	} else {
		mudlog.Info("UIModule", "config", "Using default RemoveField", "value", g.removeField)
	}
}

// GUICommand handles the 'gui' command for users
func GUICommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	// Check if this is a Mudlet client
	isMudlet := checkIfMudletClient(user.ConnectionId())
	if !isMudlet {
		user.SendText(`This command is only available for Mudlet clients.`)
		return true, nil
	}

	// Parse command arguments
	args := strings.Fields(rest)
	if len(args) == 0 {
		showGUIHelp(user)
		return true, nil
	}

	switch strings.ToLower(args[0]) {
	case "install":
		// Send GUI installation GMCP message using config values
		guiPayload := `{
			"version": "` + uiModule.uiVersion + `",
			"url": "` + uiModule.uiUrl + `"
		}`

		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `Client.GUI`,
			Payload: guiPayload,
		})

		user.SendText(`<ansi fg="command">GUI installation</ansi> package has been sent to your Mudlet client. Please follow any prompts in Mudlet to complete the installation.`)
		return true, nil

	case "remove":
		// Send GUI removal GMCP message
		removePayload := `{
			"` + uiModule.removeField + `": true
		}`

		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `Client.GUI`,
			Payload: removePayload,
		})

		user.SendText(`<ansi fg="command">GUI removal</ansi> request has been sent to your Mudlet client.`)
		return true, nil

	case "stop":
		// Add the flag to the user's config options
		user.SetConfigOption("ui_suppress_mudlet_notice", true)
		user.SendText(`The Mudlet GUI detection message has been <ansi fg="command">disabled</ansi>. You won't see it again when you log in.`)
		return true, nil

	default:
		showGUIHelp(user)
		return true, nil
	}
}

// Show help for the GUI command
func showGUIHelp(user *users.UserRecord) {
	helpText := `
<ansi fg="h1">GUI Command</ansi>

This command helps you manage the Mudlet GUI package:

<ansi fg="command">gui install</ansi> - Sends a GUI package to your Mudlet client for installation
<ansi fg="command">gui remove</ansi>  - Removes the GUI package from your Mudlet client
<ansi fg="command">gui stop</ansi>    - Stops showing the Mudlet GUI notification when you log in

The Mudlet GUI provides an enhanced interface for this MUD, with clickable widgets,
maps, and other tools to improve your gameplay experience.
`
	user.SendText(helpText)
}

// Check if the client is a Mudlet client
func checkIfMudletClient(connectionId uint64) bool {
	// Get the exported IsMudlet function from the gmcp module
	pluginReg := plugins.GetPluginRegistry()
	
	if isMudletFunc, ok := pluginReg.GetExportedFunction("IsMudlet"); ok {
		if checker, ok := isMudletFunc.(func(uint64) bool); ok {
			return checker(connectionId)
		}
	}
	
	return false
} 