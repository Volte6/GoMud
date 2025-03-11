package configs

type Scripting struct {
	LoadTimeoutMs ConfigInt `yaml:"LoadTimeoutMs"` // How long to spend the first time a script is loaded into memory
	RoomTimeoutMs ConfigInt `yaml:"RoomTimeoutMs"` // How many milliseconds to allow a script to run before it is interrupted
}

func (s *Scripting) Validate() {

	if s.LoadTimeoutMs < 1 {
		s.LoadTimeoutMs = 1000 // default
	}

	if s.RoomTimeoutMs < 1 {
		s.RoomTimeoutMs = 10
	}

}

func GetScriptingConfig() Scripting {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Scripting
}
