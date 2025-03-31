package connections

type ClientSettings struct {
	Display DisplaySettings
	Client  ClientType // Client data provided by `Core.Hello` GMCP command
	// Is MSP enabled?
	MSPEnabled        bool // Do they accept sound in their client?
	SendTelnetGoAhead bool // Defaults false, should we send a IAC GA after prompts?
}

type DisplaySettings struct {
	ScreenWidth  uint32
	ScreenHeight uint32
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
