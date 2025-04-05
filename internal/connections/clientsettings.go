package connections

type ClientSettings struct {
	Display DisplaySettings
	// Is MSP enabled?
	MSPEnabled        bool // Do they accept sound in their client?
	SendTelnetGoAhead bool // Defaults false, should we send a IAC GA after prompts?
}

type DisplaySettings struct {
	ScreenWidth  uint32
	ScreenHeight uint32
}

func (c ClientSettings) IsMsp() bool {
	return c.MSPEnabled
}
