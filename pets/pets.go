package pets

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/volte6/mud/fileloader"
	"github.com/volte6/mud/items"
	"github.com/volte6/mud/util"
	"gopkg.in/yaml.v2"
)

type Pet struct {
	Name          string         `yaml:"name,omitempty"`          // Name of the pet (player provided hopefully)
	Type          string         `yaml:"type"`                    // type of pet
	Food          Food           `yaml:"food,omitempty"`          // how much food the pet has
	LastMealRound uint8          `yaml:"lastmealround,omitempty"` // When the pet was last fed
	StatMods      map[string]int `yaml:"statmods,omitempty"`      // stat mods the pet provides
	BuffIds       []int          `yaml:"buffids,omitempty"`       // Permabuffs this pet affords the player
	Powers        []Power        `yaml:"powers,omitempty"`        // Passive abilities this pet has
	Items         []items.Item   `yaml:"items,omitempty"`         // Items held by this pet
}

var (
	petTypes = map[string]*Pet{}

	petDataFilesFolderPath = "_datafiles/pets"
)

func (p *Pet) HasPower(pwr Power) bool {
	for _, pwrName := range p.Powers {
		if pwrName == pwr {
			return true
		}
	}
	return false
}

func GetPetCopy(petId string) Pet {
	if petInfo, ok := petTypes[petId]; ok {
		return *petInfo
	}
	return Pet{}
}

func GetPetSpec(petId string) Pet {
	if petInfo, ok := petTypes[petId]; ok {
		return *petInfo
	}
	return Pet{}
}

func (p *Pet) Filename() string {
	filename := strings.ToLower(p.Type)
	filename = strings.Replace(filename, " ", "_", -1)
	return fmt.Sprintf("%s.yaml", filename)
}

func (p *Pet) Filepath() string {
	return p.Filename()
}

func (p *Pet) Save() error {
	fileName := strings.ToLower(p.Name)

	bytes, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	saveFilePath := util.FilePath(petDataFilesFolderPath, `/`, fmt.Sprintf("%s.yaml", fileName))

	err = os.WriteFile(saveFilePath, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pet) Id() string {
	return p.Type
}

func (p *Pet) Validate() error {

	if p.StatMods == nil {
		p.StatMods = map[string]int{}
	}

	if p.BuffIds == nil {
		p.BuffIds = []int{}
	}

	if p.Powers == nil {
		p.Powers = []Power{}
	}

	if p.Items == nil {
		p.Items = []items.Item{}
	}

	return nil
}

// file self loads due to init()
func LoadDataFiles() {

	start := time.Now()

	var err error
	petTypes, err = fileloader.LoadAllFlatFiles[string, *Pet](petDataFilesFolderPath)
	if err != nil {
		panic(err)
	}

	slog.Info("pets.LoadDataFiles()", "loadedCount", len(petTypes), "Time Taken", time.Since(start))
}
