package mutators

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/volte6/gomud/fileloader"
	"github.com/volte6/gomud/gametime"
	"github.com/volte6/gomud/term"
	"github.com/volte6/gomud/util"
	"gopkg.in/yaml.v2"
)

var (
	allMutators            = map[string]*MutatorSpec{}
	mutDataFilesFolderPath = "_datafiles/mutators"
)

type MutatorList []Mutator

type Mutator struct {
	MutatorId      string // Short text that will uniquely identify this modifier ("dusty")
	SpawnedRound   uint64 `yaml:"-"` // Tracks when this mutator was created (useful for decay)
	DespawnedRound uint64 `yaml:"-"` // Track when it decayed to nothing.
}

type MutatorSpec struct {
	MutatorId           string `yaml:"mutatorid,omitempty"`           // Short text that will uniquely identify this modifier ("dusty")
	NameModifier        string `yaml:"namemodifier,omitempty"`        // Text that will replace or append to existing name information (Title of a room for example) ("Dusty")
	DescriptionModifier string `yaml:"descriptionmodifier,omitempty"` // Text that will replace or append to existing descriptive information (Room description) ("The floors are covered in a thick layer of dust")
	DecayIntoId         string `yaml:"decayintoid,omitempty"`         // Id of another Mutator that replaces this one when it decays. This can be a circular behavior.
	//TODO: BuffIds             []int  // buffId's that apply conditionally (For rooms, anyone that enters the room gets the buff applied)
	DecayRate   string `yaml:"decayrate,omitempty"`   // how long until it is gone
	RespawnRate string `yaml:"respawnrate,omitempty"` // daily, weekly, 1 day, 3 day, monthly, etc.
}

func (ml *MutatorList) Add(mutName string) {

	if _, ok := allMutators[mutName]; !ok {
		return
	}

	for i, mut := range *ml {
		if mut.MutatorId == mutName {
			if !mut.Live() {
				mut.DespawnedRound = 0
				mut.SpawnedRound = 0
				mut.Update(util.GetRoundCount())
				(*ml)[i] = mut
				return
			}
		}
	}
	*ml = append(*ml, Mutator{MutatorId: mutName})
}

func (ml *MutatorList) Remove(mutName string) {

	for i, mut := range *ml {
		if mut.MutatorId == mutName {
			if mut.Live() {
				rNow := util.GetRoundCount()
				mut.DespawnedRound = rNow
				mut.SpawnedRound = 0
				(*ml)[i] = mut
				slog.Info("FOUNDMUT", "mut.DespawnedRound", mut.DespawnedRound, "mut.SpawnedRound", mut.SpawnedRound)
				(*ml).Update(rNow)
				return
			}
		}
	}
}

func (ml *MutatorList) Update(roundNow uint64) {

	if ml == nil {
		ml = &MutatorList{}
	}

	removeIdx := []int{}
	for idx := range *ml {
		(*ml)[idx].Update(roundNow)
		if (*ml)[idx].Removable() {
			removeIdx = append(removeIdx, idx)
		}
	}

	for i := len(removeIdx) - 1; i >= 0; i-- {
		(*ml) = append((*ml)[:removeIdx[i]], (*ml)[removeIdx[i]+1:]...)
	}
}

func (ml *MutatorList) NameLen() int {
	nmLen := 0
	for _, m := range *ml {
		if !m.Live() {
			continue
		}
		if m.GetSpec().NameModifier != `` {
			nmLen++
		}
	}
	return nmLen
}

func (ml *MutatorList) DescriptionLen() int {
	dLen := 0
	for _, m := range *ml {
		if !m.Live() {
			continue
		}
		if m.GetSpec().DescriptionModifier != `` {
			dLen++
		}
	}
	return dLen
}

func (ml *MutatorList) NameText() string {

	ret := strings.Builder{}
	for _, mut := range *ml {
		if !mut.Live() {
			continue
		}
		mSpec := mut.GetSpec()
		if mSpec.NameModifier == `` {
			continue
		}
		if ret.Len() > 0 {
			ret.WriteString(`, `)
		}
		ret.WriteString(mSpec.NameModifier)
	}
	return ret.String()
}

func (ml *MutatorList) DescriptionText() string {

	ret := strings.Builder{}
	for _, mut := range *ml {
		if !mut.Live() {
			continue
		}
		mSpec := mut.GetSpec()
		if mSpec.DescriptionModifier == `` {
			continue
		}
		if ret.Len() > 0 {
			ret.WriteString(term.CRLFStr)
		}
		ret.WriteString(mSpec.DescriptionModifier)
	}
	return ret.String()
}

func (m *Mutator) Live() bool {
	return m.DespawnedRound == 0
}

// Returns true if mutator can be removed since it won't become anything or respawn
func (m *Mutator) Removable() bool {
	// If currently in play, don't remove
	if m.Live() {
		return false
	}
	// If it might respawn, don't remove
	if m.GetSpec().RespawnRate != `` {
		return false
	}

	return true
}

func (m *Mutator) GetSpec() MutatorSpec {
	return *allMutators[m.MutatorId]
}

// Checks whether it decays or respawns
// Returns true if it has changed somehow?
func (m *Mutator) Update(currentRound uint64) {
	spec := m.GetSpec()

	//
	// If it hasn't been initialized yet
	//
	if m.SpawnedRound == 0 && m.DespawnedRound == 0 {
		m.SpawnedRound = currentRound
	}

	//
	// If it is currently despawned, check whether we should respawn it.
	//
	if spec.RespawnRate != `` {
		if m.DespawnedRound != 0 {
			gd := gametime.GetDate(m.DespawnedRound)
			respawnRound := gd.AddPeriod(spec.RespawnRate)

			// Has enough time passed to do the respawn?
			if currentRound >= respawnRound {
				m.DespawnedRound = 0
				m.SpawnedRound = respawnRound

			}

			return
		}
	}

	//
	// It isn't despawned, so check whether we should despawn it.
	//

	if spec.DecayRate != `` {
		gd := gametime.GetDate(m.SpawnedRound)
		despawnRound := gd.AddPeriod(spec.DecayRate)

		// Has enough time passed to do the despawn?
		if currentRound >= despawnRound {
			if spec.DecayIntoId != `` {

				m.MutatorId = spec.DecayIntoId
				m.SpawnedRound = currentRound
				m.DespawnedRound = 0

			} else {
				m.DespawnedRound = currentRound
			}

			return
		}
	}

}

func (m *MutatorSpec) Filename() string {
	filename := util.ConvertForFilename(m.MutatorId)
	return fmt.Sprintf("%s.yaml", filename)
}

func (m *MutatorSpec) Filepath() string {
	return m.Filename()
}

func (m *MutatorSpec) Save() error {
	fileName := strings.ToLower(m.MutatorId)

	bytes, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	saveFilePath := util.FilePath(mutDataFilesFolderPath, `/`, fmt.Sprintf("%s.yaml", fileName))

	err = os.WriteFile(saveFilePath, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (m *MutatorSpec) Id() string {
	return m.MutatorId
}

func (m *MutatorSpec) Validate() error {
	return nil
}

// file self loads due to init()
func LoadDataFiles() {

	start := time.Now()

	var err error
	allMutators, err = fileloader.LoadAllFlatFiles[string, *MutatorSpec](mutDataFilesFolderPath)
	if err != nil {
		panic(err)
	}

	slog.Info("mutators.LoadDataFiles()", "loadedCount", len(allMutators), "Time Taken", time.Since(start))
}
