package connections

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
	Name     string
	Version  string
	IsMudlet bool // Knowing whether is a mudlet client can be useful, since Mudlet hates certain ANSI/Escape codes.
}

// Check whether the client is Mudlet
func (c ClientSettings) IsMudlet() bool {
	return c.Client.IsMudlet
}

// Check whether a GMCP module is enabled on the client
func (c ClientSettings) GmcpEnabled(moduleName string) bool {
	if len(c.GMCPModules) == 0 {
		return false
	}

	_, ok := c.GMCPModules[moduleName]

	return ok
}
