package configs

type Integrations struct {
	Discord IntegrationsDiscord `yaml:"Discord"`
}

type IntegrationsDiscord struct {
	WebhookUrl ConfigSecret `yaml:"WebhookUrl" env:"DISCORD_WEBHOOK_URL"` // Optional Discord URL to post updates to
}

func (i *Integrations) Validate() {

	// Ignore Discord

}

func GetIntegrationsConfig() Integrations {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Integrations
}
