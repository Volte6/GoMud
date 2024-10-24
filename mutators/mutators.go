package mutators

import "github.com/volte6/gomud/gametime"

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
	// Testing
	DecayRate   string // how long until it is gone
	RespawnRate string // daily, weekly, 1 day, 3 day, monthly, etc.
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

	//
	// If it hasn't been initialized yet
	//
	if m.SpawnedRound == 0 && m.DespawnedRound == 0 {
		m.SpawnedRound = currentRound
	}

	if spec.DecayRate == `` {
		return false
	}

	//
	// If it is currently despawned, check whether we should respawn it.
	//
	if m.DespawnedRound != 0 {

		gd := gametime.GetDate(m.DespawnedRound)
		respawnRound := gd.AddPeriod(spec.RespawnRate)

		// Has enough time passed to do the respawn?
		if currentRound >= respawnRound {

			m.DespawnedRound = 0
			m.SpawnedRound = respawnRound

		}

		return true
	}

	//
	// It isn't despawned, so check whether we should despawn it.
	//
	gd := gametime.GetDate(m.SpawnedRound)
	despawnRound := gd.AddPeriod(spec.DecayRate)

	// Has enough time passed to do the despawn?
	if currentRound >= despawnRound {

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
