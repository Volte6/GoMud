package configs

type EngineTiming struct {
	TurnMs            ConfigInt `yaml:"TurnMs"`
	RoundSeconds      ConfigInt `yaml:"RoundSeconds"`
	RoundsPerAutoSave ConfigInt `yaml:"RoundsPerAutoSave"`
	RoundsPerDay      ConfigInt `yaml:"RoundsPerDay"` // How many rounds are in a day
	NightHours        ConfigInt `yaml:"NightHours"`   // How many hours of night
}

func (e *EngineTiming) Validate() {

	if e.TurnMs < 10 {
		e.TurnMs = 100 // default
	}

	if e.RoundSeconds < 1 {
		e.RoundSeconds = 4 // default
	}

	if e.RoundsPerAutoSave < 1 {
		e.RoundsPerAutoSave = 900 // default of 15 minutes worth of rounds
	}

	if e.RoundsPerDay < 10 {
		e.RoundsPerDay = 20 // default of 24 hours worth of rounds
	}

	if e.NightHours < 0 {
		e.NightHours = 0
	} else if e.NightHours > 24 {
		e.NightHours = 24
	}

}

func GetEngineTimingConfig() EngineTiming {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.EngineTiming
}
