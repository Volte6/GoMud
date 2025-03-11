package configs

type Auctions struct {
	Enabled         ConfigBool `yaml:"Enabled"`
	Anonymous       ConfigBool `yaml:"Anonymous"`
	DurationSeconds ConfigInt  `yaml:"DurationSeconds"`
	UpdateSeconds   ConfigInt  `yaml:"UpdateSeconds"`
}

func (a *Auctions) Validate() {

	// Ignore Enabled
	// Ignore Anonymous

	if a.DurationSeconds < 30 {
		a.DurationSeconds = 30 // minimum
	}

	if a.UpdateSeconds < 15 {
		a.UpdateSeconds = 15 // default
	} else if a.UpdateSeconds > a.UpdateSeconds>>1 {
		a.UpdateSeconds = a.UpdateSeconds >> 1 // default
	}

}

func GetAuctionsConfig() Auctions {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Auctions
}
