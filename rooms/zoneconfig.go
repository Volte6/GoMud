package rooms

type ZoneConfig struct {
	RoomId       int `yaml:"roomid,omitempty"`
	MobAutoScale struct {
		Minimum int `yaml:"minimum,omitempty"` // level scaling minimum
		Maximum int `yaml:"maximum,omitempty"` // level scaling maximum
	} `yaml:"autoscale,omitempty"` // level scaling range if any
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
