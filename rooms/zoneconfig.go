package rooms

import (
	"github.com/volte6/gomud/mutators"
	"github.com/volte6/gomud/util"
)

type ZoneConfig struct {
	RoomId       int `yaml:"roomid,omitempty"`
	MobAutoScale struct {
		Minimum int `yaml:"minimum,omitempty"` // level scaling minimum
		Maximum int `yaml:"maximum,omitempty"` // level scaling maximum
	} `yaml:"autoscale,omitempty"` // level scaling range if any
	SpawnCooldown int                  `yaml:"spawncooldown,omitempty"` // default cooldown if no other specified
	Mutators      mutators.MutatorList `yaml:"mutators,omitempty"`      // mutators defined here apply to entire zone
}

func (z *ZoneConfig) Validate() {
	if z.MobAutoScale.Minimum < 0 {
		z.MobAutoScale.Minimum = 0
	}

	if z.MobAutoScale.Maximum < 0 {
		z.MobAutoScale.Maximum = 0
	}

	// If either is set, neither can be zero.
	if z.MobAutoScale.Minimum > 0 || z.MobAutoScale.Maximum > 0 {

		if z.MobAutoScale.Maximum < z.MobAutoScale.Minimum {
			z.MobAutoScale.Maximum = z.MobAutoScale.Minimum
		}

		if z.MobAutoScale.Minimum == 0 {
			z.MobAutoScale.Minimum = z.MobAutoScale.Maximum
		}
	}

	if z.SpawnCooldown < 0 {
		z.SpawnCooldown = 0
	}
}

// Generates a random number between min and max
func (z *ZoneConfig) GenerateRandomLevel() int {
	return util.Rand(z.MobAutoScale.Maximum-z.MobAutoScale.Minimum) + z.MobAutoScale.Minimum
}
