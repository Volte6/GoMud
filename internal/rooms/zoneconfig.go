package rooms

import (
	"github.com/GoMudEngine/GoMud/internal/mutators"
	"github.com/GoMudEngine/GoMud/internal/util"
)

type ZoneConfig struct {
	RoomId       int `yaml:"roomid,omitempty"`
	MobAutoScale struct {
		Minimum int `yaml:"minimum,omitempty"` // level scaling minimum
		Maximum int `yaml:"maximum,omitempty"` // level scaling maximum
	} `yaml:"autoscale,omitempty"` // level scaling range if any
	Mutators     mutators.MutatorList `yaml:"mutators,omitempty"`     // mutators defined here apply to entire zone
	IdleMessages []string             `yaml:"idlemessages,omitempty"` // list of messages that can be displayed to players in the zone, assuming a room has none defined
	MusicFile    string               `yaml:"musicfile,omitempty"`    // background music to play when in this zone
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

}

// Generates a random number between min and max
func (z *ZoneConfig) GenerateRandomLevel() int {
	return util.Rand(z.MobAutoScale.Maximum-z.MobAutoScale.Minimum) + z.MobAutoScale.Minimum
}
