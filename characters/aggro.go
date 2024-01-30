package characters

type AggroType int

const (
	// Enumerated Aggro Types
	DefaultAttack AggroType = iota // Regular H2H combat, everything can decay to this. Starts at zero
	Shooting
	BackStab
	SpellCast
	Flee
	Aid
)

type Aggro struct {
	Type          AggroType
	MobInstanceId int
	UserId        int
	SpellName     string // If Type is SpellCast, this is the spell name
	ExitName      string // For example, firing a weapon in a direction
	RoundsWaiting int    // How many rounds must pass before this triggers
}
