package configs

type Modules map[string]any

func (p *Modules) Validate() {

}

func GetModulesConfig() Modules {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Modules
}
