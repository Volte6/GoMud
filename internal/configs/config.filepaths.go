package configs

type FilePaths struct {
	WebCDNLocation   ConfigString `yaml:"WebCDNLocation"`
	FolderDataFiles  ConfigString `yaml:"FolderDataFiles"`
	FolderPublicHtml ConfigString `yaml:"FolderPublicHtml"`
	FolderAdminHtml  ConfigString `yaml:"FolderAdminHtml"`
	HttpsCertFile    ConfigString `yaml:"HttpsCertFile"`
	HttpsKeyFile     ConfigString `yaml:"HttpsKeyFile"`
	CarefulSaveFiles ConfigBool   `yaml:"CarefulSaveFiles"`
}

func (f *FilePaths) Validate() {

	// Ignore WebCDNLocation
	// Ignore FolderPublicHtml
	// Ignore FolderAdminHtml
	// Ignore CarefulSaveFiles

	if f.FolderDataFiles == `` {
		f.FolderDataFiles = `_datafiles/world/default` // default
	}

}

func GetFilePathsConfig() FilePaths {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.FilePaths
}
