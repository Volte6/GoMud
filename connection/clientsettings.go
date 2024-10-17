package connection

type ClientSettings struct {
	Display DisplaySettings
	Discord DiscordSettings
	Client  ClientType
	// Enabled GMCP Modules
	GMCPModules map[string]int
}

type DisplaySettings struct {
	ScreenWidth  uint32
	ScreenHeight uint32
}
type DiscordSettings struct {
	User    string // person#1234
	Private bool
}

type ClientType struct {
	Name    string
	Version string
}
