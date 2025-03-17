package configs

type Statistics struct {
	Leaderboards StatisticsLeaderboards `yaml:"Leaderboards"` // Maximum size of leaderboard
}

type StatisticsLeaderboards struct {
	Size              ConfigInt  `yaml:"Size"`              // Maximum size of leaderboard
	ExperienceEnabled ConfigBool `yaml:"ExperienceEnabled"` // Enable XP leaderboards?
	GoldEnabled       ConfigBool `yaml:"GoldEnabled"`       // Enable Gold leaderboards?
	KillsEnabled      ConfigBool `yaml:"KillsEnabled"`      // Enable Kills leaderboards?
}

func (s *Statistics) Validate() {

	if s.Leaderboards.Size < 0 {
		s.Leaderboards.Size = 0
	}

}

func GetStatisticsConfig() Statistics {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Statistics
}
