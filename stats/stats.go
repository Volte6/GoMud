package stats

const (
	BaseModFactor         = 0.3333333334 // How much of a scaling to aply to levels before multiplying by racial stat
	NaturalGainsModFactor = 0.5          // Free stats gained per level modded by this.
)

type Statistics struct {
	Strength   StatInfo `yaml:"strength,omitempty"`   // Muscular strength (damage?)
	Speed      StatInfo `yaml:"speed,omitempty"`      // Speed and agility (dodging)
	Smarts     StatInfo `yaml:"smarts,omitempty"`     // Intelligence and wisdom (magic power, memory, deduction, etc)
	Vitality   StatInfo `yaml:"vitality,omitempty"`   // Health and stamina (health capacity)
	Mysticism  StatInfo `yaml:"mysticism,omitempty"`  // Magic and mana (magic capacity)
	Perception StatInfo `yaml:"perception,omitempty"` // How well you notice things
}

// When saving to a file, we don't need to write all the properties that we calculate.
// Just keep track of "Training" because that's not calculated.
type StatInfo struct {
	Training int `yaml:"training,omitempty"` // How much it's been trained with Training Points spending
	Value    int `yaml:"-"`                  // Final calculated value
	Base     int `yaml:"base,omitempty"`     // Base stat value
	Mods     int `yaml:"-"`                  // How much it's modded by equipment, spells, etc.
}

func (si *StatInfo) SetMod(mod ...int) {
	if len(mod) == 0 {
		si.Mods = 0
		return
	}
	si.Mods = 0
	for _, m := range mod {
		si.Mods += m
	}
}

func (si *StatInfo) GainsForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	levelScale := float64(level-1) * BaseModFactor
	basePoints := int(levelScale * float64(si.Base))

	// every x levels we get natural gains
	freeStatPoints := int(float64(level) * NaturalGainsModFactor)

	return basePoints + freeStatPoints
}

func (si *StatInfo) Recalculate(level int) {
	si.Value = si.GainsForLevel(level) + si.Training + si.Mods
}
