package configs

type Roles map[string]ConfigSliceString

func (m *Roles) Validate() {
}

func GetRolesConfig() Roles {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}

	return configData.Roles
}
