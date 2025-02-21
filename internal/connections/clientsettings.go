package connections

type ClientSettings struct {
	Display DisplaySettings
	Discord DiscordSettings
	Client  ClientType // Client data provided by `Core.Hello` GMCP command
	// Enabled GMCP Modules
	GMCPModules map[string]int // Enabled GMCP Modules (if any)
	// Is MSP enabled?
	MSPEnabled        bool // Do they accept sound in their client?
	SendTelnetGoAhead bool // Defaults false, should we send a IAC GA after prompts?
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

func (c ClientSettings) IsMsp() bool {
	return c.MSPEnabled
}

// Check whether a GMCP module is enabled on the client
func (c ClientSettings) GmcpEnabled(moduleName string) bool {
	if len(c.GMCPModules) == 0 {
		return false
	}

	_, ok := c.GMCPModules[moduleName]

	return ok
}
