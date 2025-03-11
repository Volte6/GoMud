package buffs

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/fileloader"
	"github.com/volte6/gomud/internal/gametime"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/statmods"
	"github.com/volte6/gomud/internal/util"
)

// Something temporarily attached to a character
// That modifies some aspect of their status
/*
Examples:
Fast Healing - increased natural health recovery for 10 rounds
Poison - add -10 health every round for 5 rounds
*/

type Flag string

const (
	//
	// All Flags must be lowercase
	//
	All Flag = ``

	// Behavioral flags
	NoCombat       Flag = `no-combat`
	NoMovement     Flag = `no-go`
	NoFlee         Flag = `no-flee`
	CancelIfCombat Flag = `cancel-on-combat`
	CancelOnAction Flag = `cancel-on-action`
	CancelOnWater  Flag = `cancel-on-water`

	// Death preventing
	ReviveOnDeath Flag = `revive-on-death`

	// Gear related
	PermaGear   Flag = `perma-gear`
	RemoveCurse Flag = `remove-curse`

	// Harmful flags
	Poison Flag = `poison`
	Drunk  Flag = `drunk`

	// Useful flags
	Hidden       Flag = `hidden`
	Accuracy     Flag = `accuracy`
	Blink        Flag = `blink`
	EmitsLight   Flag = `lightsource`
	SuperHearing Flag = `superhearing`
	NightVision  Flag = `nightvision`
	Warmed       Flag = `warmed`
	Hydrated     Flag = `hydrated`
	Thirsty      Flag = `thirsty`

	// Flags that reveal things
	SeeHidden Flag = `see-hidden`
	SeeNouns  Flag = `see-nouns`

	// Arbitrarily chosen round for calculating trigger round counts
	validationRound = 1000000
)

var (
	buffs map[int]*BuffSpec = make(map[int]*BuffSpec)

	validationCalculator = gametime.GetDate(validationRound)
)

type BuffSpec struct {
	BuffId        int               // Unique identifier for this buff spec
	Name          string            // The name of the buff
	Description   string            // A description of the buff
	Secret        bool              // Whether or not the buff is secret (not displayed to the user)
	TriggerNow    bool              `yaml:"triggernow,omitempty"`    // if true, buff triggers once right when it is applied
	TriggerRate   string            `yaml:"triggerrate,omitempty"`   // How often should it trigger? (time string)
	RoundInterval int               `yaml:"roundinterval,omitempty"` // triggers every x rounds
	TriggerCount  int               `yaml:"triggercount,omitempty"`  // How many times it triggers before it is removed
	StatMods      statmods.StatMods `yaml:"statmods,omitempty"`      // stat mods for the duration of the buff
	Flags         []Flag            `yaml:"flags,omitempty"`         // A list of actions and such that this buff prevents or enables
}

// Calculates the value of this buff
func (b *BuffSpec) GetValue() int {
	val := 0

	for _, v := range b.StatMods {
		val += int(math.Abs(float64(v)))
	}

	freqVal := 5 - b.RoundInterval
	if freqVal < 0 {
		freqVal = 0
	}
	val += freqVal
	val += len(b.Flags) * 5

	if b.TriggerCount > 0 {
		val *= b.TriggerCount
	}

	return val
}

type BuffMessage struct {
	User string
	Room string
}

type BuffMessages struct {
	Start  BuffMessage
	Effect BuffMessage
	End    BuffMessage
}

func GetBuffSpec(buffId int) *BuffSpec {
	if buffId < 0 {
		buffId *= -1
	}

	if buff, ok := buffs[buffId]; ok {
		return buff
	}

	return nil
}

func GetAllBuffIds() []int {

	var results []int = make([]int, 0, len(buffs))
	for _, buff := range buffs {
		results = append(results, buff.BuffId)
	}

	return results
}

// Searches for buffs whos name contain text and returns thehr Ids
func SearchBuffs(searchTerm string) []int {

	searchTerm = strings.TrimSpace(strings.ToLower(searchTerm))

	var results []int = make([]int, 0, 2)

	for _, buff := range buffs {
		if strings.Contains(strings.ToLower(buff.Name), searchTerm) {
			results = append(results, buff.BuffId)
		} else if strings.Contains(strings.ToLower(buff.Description), searchTerm) {
			results = append(results, buff.BuffId)
		}
	}

	return results
}

// Presumably to ensure the datafile hasn't messed something up.
func (b *BuffSpec) Id() int {
	return b.BuffId
}

// Presumably to ensure the datafile hasn't messed something up.
func (b *BuffSpec) Validate() error {

	b.RoundInterval = int(validationCalculator.AddPeriod(b.TriggerRate) - validationRound)

	if b.TriggerCount < 1 {
		return fmt.Errorf("buffId %d (%s) has a TriggersCount of < 1, must be at least 1", b.BuffId, b.Name)
	}
	if b.RoundInterval < 1 {
		return fmt.Errorf("buffId %d (%s) has a RoundInterval of < 1, must be at least 1. Is %s a valid time string?", b.BuffId, b.Name, b.TriggerRate)
	}
	return nil
}

func (b *BuffSpec) Filename() string {
	filename := util.ConvertForFilename(b.Name)
	return fmt.Sprintf("%d-%s.yaml", b.BuffId, filename)
}

func (b *BuffSpec) Filepath() string {
	return b.Filename()
}

func (b *BuffSpec) GetScript() string {

	scriptPath := b.GetScriptPath()
	// Load the script into a string
	if _, err := os.Stat(scriptPath); err == nil {
		if bytes, err := os.ReadFile(scriptPath); err == nil {
			return string(bytes)
		}
	}

	return ``
}

func (b *BuffSpec) GetScriptPath() string {
	// Load any script for the buff

	buffFilePath := b.Filename()
	scriptFilePath := strings.Replace(buffFilePath, `.yaml`, `.js`, 1)

	fullScriptPath := strings.Replace(string(configs.GetFilePathsConfig().FolderDataFiles)+`/buffs/`+b.Filepath(),
		buffFilePath,
		scriptFilePath,
		1)

	return util.FilePath(fullScriptPath)
}

// file self loads due to init()
func LoadDataFiles() {

	start := time.Now()

	tmpBuffs, err := fileloader.LoadAllFlatFiles[int, *BuffSpec](string(configs.GetFilePathsConfig().FolderDataFiles) + `/buffs`)
	if err != nil {
		panic(err)
	}

	buffs = tmpBuffs

	mudlog.Info("buffSpec.LoadDataFiles()", "loadedCount", len(buffs), "Time Taken", time.Since(start))
}
