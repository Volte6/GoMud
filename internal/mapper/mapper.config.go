package mapper

type Config struct {
	ZoomLevel       int                    // How much to zoom the map? At zero, connections are not shown.
	Width           int                    // Width of final map render
	Height          int                    // Height of final map render
	UserId          int                    // Optional userId to intelligently render some data
	symbolOverrides map[int]SymbolOverride // Force symbols/legends for a specific room
}

type SymbolOverride struct {
	Symbol rune
	Legend string
}

func (c *Config) OverrideSymbol(roomId int, symbol rune, legend string) {
	if c.symbolOverrides == nil {
		c.symbolOverrides = make(map[int]SymbolOverride)
	}

	c.symbolOverrides[roomId] = SymbolOverride{symbol, legend}
}
