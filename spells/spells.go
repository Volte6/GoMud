package spells

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/fileloader"
	"github.com/volte6/mud/util"
)

type SpellType string

type SpellData struct {
	SpellId     string
	Name        string
	Description string
	Type        SpellType
	Cost        int
	WaitRounds  int
	Difficulty  int // Augments final success chance by this %
}

const (
	WaitRoundsDefault = 3

	Neutral    SpellType = "neutral"    // Neutral, no expected actor target, use on
	HarmSingle SpellType = "harmsingle" // Harmful, defaults to current aggro - magic missile etc
	HarmMulti  SpellType = "harmmulti"  // Harmful, defaults to all aggro mobs - chain lightning etc
	HelpSingle SpellType = "helpsingle" // Helpful, defaults on self - heal etc
	HelpMulti  SpellType = "helpmulti"  // Helpful, defaults on party - mass heal etc
)

var (
	allSpells = map[string]*SpellData{}
)

func (s SpellType) HelpOrHarmString() string {
	switch s {
	case Neutral:
		return `Neutral`
	case HelpSingle:
		return `Helpful`
	case HelpMulti:
		return `Helpful`
	case HarmSingle:
		return `Harmful`
	case HarmMulti:
		return `Harmful`
	}
	return `Unknown`
}

func (s SpellType) TargetTypeString() string {
	switch s {
	case Neutral:
		return `Unknown`
	case HelpSingle:
		return `Single Target`
	case HarmSingle:
		return `Single Target`
	case HelpMulti:
		return `Group Target`
	case HarmMulti:
		return `Group Target`
	}
	return `Unknown`
}

func GetSpell(spellId string) *SpellData {
	if sd, ok := allSpells[spellId]; ok {
		return sd
	}
	return nil
}

func FindSpellByName(spellName string) *SpellData {

	var closestMatch *SpellData = nil

	spellName = strings.ToLower(spellName)
	for _, spellData := range allSpells {

		testName := strings.ToLower(spellData.Name)

		if testName == spellName {
			return spellData
		}

		if closestMatch == nil && strings.HasPrefix(strings.ToLower(spellData.Name), spellName) {
			closestMatch = spellData
		}

	}
	return closestMatch
}

func GetAllSpells() map[string]*SpellData {
	retSpellBook := make(map[string]*SpellData)
	for k, v := range allSpells {
		retSpellBook[k] = v
	}
	return retSpellBook
}

func (s *SpellData) Id() string {
	return s.SpellId
}

// SpellData implements the Filepath method from the Loadable interface.
func (s *SpellData) Filepath() string {
	return util.FilePath(fmt.Sprintf("%s.yaml", s.SpellId))
}

func (s *SpellData) Validate() error {
	return nil
}

func (s *SpellData) GetScript() string {

	scriptPath := s.GetScriptPath()

	// Load the script into a string
	if _, err := os.Stat(scriptPath); err == nil {
		if bytes, err := os.ReadFile(scriptPath); err == nil {
			return string(bytes)
		}
	}

	return ``
}

func (s *SpellData) GetScriptPath() string {
	// Load any script for the room
	return strings.Replace(string(configs.GetConfig().FolderSpellData)+`/`+s.Filepath(), `.yaml`, `.js`, 1)
}

func LoadSpellFiles() {

	start := time.Now()

	var err error
	allSpells, err = fileloader.LoadAllFlatFiles[string, *SpellData](string(configs.GetConfig().FolderSpellData))
	if err != nil {
		panic(err)
	}

	slog.Info("spells.loadAllSpells()", "loadedCount", len(allSpells), "Time Taken", time.Since(start))

}
