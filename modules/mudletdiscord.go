package modules

import (
	"embed"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var (
	//go:embed mudletdiscord/*
	mudletdiscord_Files embed.FS
)

func init() {
	// Create the module instance first
	g := MudletDiscordModule{
		plug: plugins.New(`mudletdiscord`, `1.0`),
	}

	// Attach the file system to the plugin
	if err := g.plug.AttachFileSystem(mudletdiscord_Files); err != nil {
		panic(err)
	}

	// Set default values directly during initialization
	g.setDefaultValues()

	// Register the onLoad callback to load config values after plugin system is ready
	g.plug.Callbacks.SetOnLoad(func() {
		g.RefreshConfig()
	})

	// Export a function to register event listeners for any user with GMCP support
	// The function name remains "RegisterMudletUser" for backward compatibility
	g.plug.ExportFunction("RegisterMudletUser", g.registerEventListeners)
}

type MudletDiscordModule struct {
	plug          *plugins.Plugin
	inviteUrl     string
	applicationId string
	largeImageKey string
	state         string
	details       string
	mudName       string
	// Track which users already have event listeners registered
	registeredUsers map[int]bool
}

// Set hardcoded default values for initialization
func (g *MudletDiscordModule) setDefaultValues() {
	g.mudName = "GoMud"
	g.inviteUrl = "https://discord.gg/FaauSYej3n"
	g.applicationId = "1234"
	g.largeImageKey = "server-icon"
	g.state = "Playing GoMud"
	g.details = "using GoMud Engine"
	g.registeredUsers = make(map[int]bool)
}

// This function gets called when a client with GMCP support is detected
// Function name kept as is for backward compatibility
func (g *MudletDiscordModule) registerEventListeners(userId int) bool {
	// Check if we've already registered this user
	if _, ok := g.registeredUsers[userId]; ok {
		return true // Already registered
	}
	
	// Register the user and send initial GMCP data
	g.registeredUsers[userId] = true
	
	// Get the user record
	user := users.GetByUserId(userId)
	if user == nil {
		return false
	}
	
	// Send initial Discord info
	g.sendDiscordInfo(user)
	
	return true
}

// RefreshConfig loads values from the module's config.yaml file
func (g *MudletDiscordModule) RefreshConfig() {
	// Using direct config properties like auctions.go does
	if inviteUrl, ok := g.plug.Config.Get(`InviteUrl`).(string); ok && inviteUrl != "" {
		g.inviteUrl = inviteUrl
		mudlog.Info("MudletDiscordModule", "config", "Loaded InviteUrl from config", "value", g.inviteUrl)
	} else {
		mudlog.Info("MudletDiscordModule", "config", "Using default InviteUrl", "value", g.inviteUrl)
	}

	if applicationId, ok := g.plug.Config.Get(`ApplicationId`).(string); ok && applicationId != "" {
		g.applicationId = applicationId
		mudlog.Info("MudletDiscordModule", "config", "Loaded ApplicationId from config", "value", g.applicationId)
	} else {
		mudlog.Info("MudletDiscordModule", "config", "Using default ApplicationId", "value", g.applicationId)
	}

	if largeImageKey, ok := g.plug.Config.Get(`LargeImageKey`).(string); ok && largeImageKey != "" {
		g.largeImageKey = largeImageKey
		mudlog.Info("MudletDiscordModule", "config", "Loaded LargeImageKey from config", "value", g.largeImageKey)
	} else {
		mudlog.Info("MudletDiscordModule", "config", "Using default LargeImageKey", "value", g.largeImageKey)
	}

	if state, ok := g.plug.Config.Get(`State`).(string); ok && state != "" {
		g.state = state
		mudlog.Info("MudletDiscordModule", "config", "Loaded State from config", "value", g.state)
	} else {
		mudlog.Info("MudletDiscordModule", "config", "Using default State", "value", g.state)
	}

	if details, ok := g.plug.Config.Get(`Details`).(string); ok && details != "" {
		g.details = details
		mudlog.Info("MudletDiscordModule", "config", "Loaded Details from config", "value", g.details)
	} else {
		mudlog.Info("MudletDiscordModule", "config", "Using default Details", "value", g.details)
	}

	// Always get MudName from main config
	c := configs.GetConfig()
	if c.Server.MudName != "" {
		g.mudName = string(c.Server.MudName)
		mudlog.Info("MudletDiscordModule", "config", "Loaded MudName from main config", "value", g.mudName)
	} else {
		mudlog.Info("MudletDiscordModule", "config", "Using default MudName", "value", g.mudName)
	}
}

// Helper method to send Discord info to a user
func (g *MudletDiscordModule) sendDiscordInfo(user *users.UserRecord) {
	if g.inviteUrl != "" {
		infoPayload := `{ 
			"inviteurl": "` + g.inviteUrl + `",	
			"applicationid": "` + g.applicationId + `",
			"largeImageKey": "` + g.largeImageKey + `"
		}`

		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `External.Discord.Info`,
			Payload: infoPayload,
		})

		// Send Discord Status
		statusPayload := `{ 
			"game": "` + g.mudName + `",
			"startTimestamp": ` + strconv.FormatInt(user.GetConnectTime().Unix(), 10) + `,
			"state": "` + g.state + `",
			"details": "` + g.details + `"
		}`

		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `External.Discord.Status`,
			Payload: statusPayload,
		})
	}
}
