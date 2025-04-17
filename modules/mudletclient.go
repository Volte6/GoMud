package modules

import (
	"embed"
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/users"
)

var (
	//go:embed mudletclient/*
	mudletclient_Files embed.FS
	
	// Instance accessible from elsewhere in this package
	mudletClientModule MudletClientModule
)

func init() {
	// Create the module instance first
	mudletClientModule = MudletClientModule{
		plug: plugins.New(`mudletclient`, `1.0`),
	}

	// Attach the file system to the plugin
	if err := mudletClientModule.plug.AttachFileSystem(mudletclient_Files); err != nil {
		panic(err)
	}

	// Set default values directly during initialization
	mudletClientModule.setDefaultValues()

	// Register the onLoad callback to load config values after plugin system is ready
	mudletClientModule.plug.Callbacks.SetOnLoad(func() {
		mudletClientModule.RefreshConfig()
	})

	// Export a function to register client info for Mudlet clients
	mudletClientModule.plug.ExportFunction("SendClientInfo", mudletClientModule.SendClientInfoExported)
}

type MudletClientModule struct {
	plug          *plugins.Plugin
	uiUrl         string
	uiVersion     string
	mapUrl        string
	mapperUrl     string
	mapperVersion string
	// Track which users already have data sent
	registeredUsers map[int]bool
}

// Set hardcoded default values for initialization
func (g *MudletClientModule) setDefaultValues() {
	g.uiUrl = ""
	g.uiVersion = "1.0.0"
	g.mapUrl = "https://github.com/MorquinDevlar/GoMudUI/releases/latest/download/gomud.dat"
	g.mapperUrl = "https://github.com/MorquinDevlar/GoMudUI/releases/latest/download/mudlet-mapper.mpackage"
	g.mapperVersion = "1.0.0"
	g.registeredUsers = make(map[int]bool)
}

// Public function that can be called from the GMCP module
func (g *MudletClientModule) SendClientInfoExported(userId int) bool {
	user := users.GetByUserId(userId)
	if user == nil {
		return false
	}
	
	// Check if we've already sent info to this user
	if _, ok := g.registeredUsers[userId]; ok {
		return true // Already sent
	}
	
	// Mark as registered
	g.registeredUsers[userId] = true
	
	// Send initial client info
	return g.sendClientInfo(user)
}

// RefreshConfig loads values from the module's config.yaml file
func (g *MudletClientModule) RefreshConfig() {
	// Load config values from the plugin's config file
	if uiUrl, ok := g.plug.Config.Get(`Mudlet.UIUrl`).(string); ok && uiUrl != "" {
		g.uiUrl = uiUrl
	}

	if uiVersion, ok := g.plug.Config.Get(`Mudlet.UIVersion`).(string); ok && uiVersion != "" {
		g.uiVersion = uiVersion
	}

	if mapUrl, ok := g.plug.Config.Get(`Mudlet.MapUrl`).(string); ok && mapUrl != "" {
		g.mapUrl = mapUrl
	}

	if mapperUrl, ok := g.plug.Config.Get(`Mudlet.MapperUrl`).(string); ok && mapperUrl != "" {
		g.mapperUrl = mapperUrl
	}

	if mapperVersion, ok := g.plug.Config.Get(`Mudlet.MapperVersion`).(string); ok && mapperVersion != "" {
		g.mapperVersion = mapperVersion
	}
}

// Helper method to send client info to a user
func (g *MudletClientModule) sendClientInfo(user *users.UserRecord) bool {
	anyInfoSent := false
	
	// Send GUI URL if configured
	if g.uiUrl != "" {
		guiPayload := fmt.Sprintf(`{ 
			"version": "%s",
			"url": "%s"
		}`, g.uiVersion, g.uiUrl)

		// Log specific module being sent
		mudlog.Info("GMCP", "action", "Sending Client.GUI to Mudlet user", "userId", user.UserId)
		
		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `Client.GUI`,
			Payload: guiPayload,
		})
		anyInfoSent = true
	}

	// Send Map URL if configured
	if g.mapUrl != "" {
		mapPayload := fmt.Sprintf(`{ 
			"url": "%s"
		}`, g.mapUrl)

		// Log specific module being sent
		mudlog.Info("GMCP", "action", "Sending Client.Map to Mudlet user", "userId", user.UserId)
		
		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `Client.Map`,
			Payload: mapPayload,
		})
		anyInfoSent = true
	}

	// Send Mapper URL if configured
	if g.mapperUrl != "" {
		mapperPayload := fmt.Sprintf(`{ 
			"version": "%s",
			"url": "%s"
		}`, g.mapperVersion, g.mapperUrl)

		// Log specific module being sent
		mudlog.Info("GMCP", "action", "Sending Client.GUI mapper info to Mudlet user", "userId", user.UserId)
		
		events.AddToQueue(GMCPOut{
			UserId:  user.UserId,
			Module:  `Client.GUI`,
			Payload: mapperPayload,
		})
		anyInfoSent = true
	}
	
	return anyInfoSent
}