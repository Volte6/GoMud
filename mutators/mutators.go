package mutators

var (
	allMutators = map[string]*MutatorSpec{}
)

type Mutator struct {
	Id             string // Short text that will uniquely identify this modifier ("dusty")
	SpawnedRound   uint64 // Tracks when this mutator was created (useful for decay)
	DespawnedRound uint64 // Track when it decayed to nothing.
}

type MutatorSpec struct {
	Id                  string // Short text that will uniquely identify this modifier ("dusty")
	NameModifier        string // Text that will replace or append to existing name information (Title of a room for example) ("Dusty")
	DescriptionModifier string // Text that will replace or append to existing descriptive information (Room description) ("The floors are covered in a thick layer of dust")
	SpawnRate           string // X hours/days/weeks/years
	DecayRounds         uint64 // how many rounds until this despawns. 0 means doesn't decay.
	DecayIntoId         string // Id of another Mutator that replaces this one when it decays. This can be a circular behavior.
	BuffIds             []int  // buffId's that apply conditionally (For rooms, anyone that enters the room gets the buff applied)
	SpawnedRound        uint64 // tracks when this mutator was created (useful for decay)
}

func (m *Mutator) Live() bool {
	return m.DespawnedRound == 0
}

func (m *Mutator) GetSpec() MutatorSpec {
	return *allMutators[m.Id]
}

// Checks whether it decays or respawns
func (m *Mutator) Update(currentRound uint64) bool {
	spec := m.GetSpec()

	// If it hasn't been initialized yet
	if m.SpawnedRound == 0 && m.DespawnedRound == 0 {
		m.SpawnedRound = currentRound
	}

	if spec.DecayRounds == 0 {
		return false
	}

	if currentRound-m.SpawnedRound >= spec.DecayRounds {

		if spec.DecayIntoId != `` {

			m.Id = spec.DecayIntoId
			m.SpawnedRound = currentRound
			m.DespawnedRound = 0

		} else {
			m.DespawnedRound = currentRound
		}

		return true
	}

	return false
}
