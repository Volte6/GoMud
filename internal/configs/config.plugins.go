package configs

type Plugins map[string]any

func (p *Plugins) Validate() {

}

func GetPluginConfig() Plugins {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Plugins
}
