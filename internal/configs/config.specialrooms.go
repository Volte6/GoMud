package configs

type SpecialRooms struct {
	StartRoom          ConfigInt         `yaml:"StartRoom"`          // Default starting room.
	DeathRecoveryRoom  ConfigInt         `yaml:"DeathRecoveryRoom"`  // Recovery room after dying.
	TutorialStartRooms ConfigSliceString `yaml:"TutorialStartRooms"` // List of all rooms that can be used to begin the tutorial process
}

func (s *SpecialRooms) Validate() {

	// Ignore StartRoom
	// Ignore DeathRecoveryRoom
	// Ignore TutorialStartRooms

}

func GetSpecialRoomsConfig() SpecialRooms {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.SpecialRooms
}
