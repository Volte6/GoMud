package configs

type Statistics struct {
	LeaderboardSize ConfigInt `yaml:"LeaderboardSize"` // Maximum size of leaderboard
}

func (s *Statistics) Validate() {

	if s.LeaderboardSize < 0 {
		s.LeaderboardSize = 0
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
