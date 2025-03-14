package mapper

// represents a single room
type mapNode struct {
	RoomId      int
	Symbol      rune
	Legend      string // The same that shows in the legend for this symbol
	Exits       map[string]nodeExit
	SecretExits map[string]struct{} // Just a flag for whether an exit key is secret
	Pos         positionDelta       // Its x/y/z position relative to the root node
}

type nodeExit struct {
	RoomId    int  // where it leads to
	Secret    bool // is it secret?
	Locked    bool // is it locked?
	Direction positionDelta
}
