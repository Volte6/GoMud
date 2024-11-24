package configs

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/volte6/gomud/internal/util"
	"gopkg.in/yaml.v2"
)

const defaultConfigPath = "_datafiles/config.yaml"

type Config struct {
	Version                      ConfigString      `yaml:"Version"` // Cuurrent version of all datafiles
	MaxCPUCores                  ConfigInt         `yaml:"MaxCPUCores"`
	FolderItemData               ConfigString      `yaml:"FolderItemData"`
	FolderAttackMessageData      ConfigString      `yaml:"FolderAttackMessageData"`
	FolderUserData               ConfigString      `yaml:"FolderUserData"`
	FolderSpellData              ConfigString      `yaml:"FolderSpellData"`
	FolderTemplates              ConfigString      `yaml:"FolderTemplates"`
	FileAnsiAliases              ConfigString      `yaml:"FileAnsiAliases"`
	FileKeywords                 ConfigString      `yaml:"FileKeywords"`
	AllowItemBuffRemoval         ConfigBool        `yaml:"AllowItemBuffRemoval"`
	CarefulSaveFiles             ConfigBool        `yaml:"CarefulSaveFiles"`
	AuctionsEnabled              ConfigBool        `yaml:"AuctionsEnabled"`
	AuctionsAnonymous            ConfigBool        `yaml:"AuctionsAnonymous"`
	AuctionSeconds               ConfigInt         `yaml:"AuctionSeconds"`
	AuctionUpdateSeconds         ConfigInt         `yaml:"AuctionUpdateSeconds"`
	PVPEnabled                   ConfigBool        `yaml:"PVPEnabled"`
	XPScale                      ConfigFloat       `yaml:"XPScale"`
	TurnMs                       ConfigInt         `yaml:"TurnMs"`
	RoundSeconds                 ConfigInt         `yaml:"RoundSeconds"`
	RoundsPerAutoSave            ConfigInt         `yaml:"RoundsPerAutoSave"`
	RoundsPerDay                 ConfigInt         `yaml:"RoundsPerDay"` // How many rounds are in a day
	NightHours                   ConfigInt         `yaml:"NightHours"`   // How many hours of night
	MaxMobBoredom                ConfigInt         `yaml:"MaxMobBoredom"`
	ScriptLoadTimeoutMs          ConfigInt         `yaml:"ScriptLoadTimeoutMs"`          // How long to spend the first time a script is loaded into memory
	ScriptRoomTimeoutMs          ConfigInt         `yaml:"ScriptRoomTimeoutMs"`          // How many milliseconds to allow a script to run before it is interrupted
	MaxTelnetConnections         ConfigInt         `yaml:"MaxTelnetConnections"`         // Maximum number of telnet connections to accept
	TelnetPort                   ConfigSliceString `yaml:"TelnetPort"`                   // One or more Ports used to accept telnet connections
	LocalPort                    ConfigInt         `yaml:"LocalPort"`                    // Port used for admin connections, localhost only
	WebPort                      ConfigInt         `yaml:"WebPort"`                      // Port used for web requests
	NextRoomId                   ConfigInt         `yaml:"NextRoomId"`                   // The next room id to use when creating a new room
	LootGoblinRoundCount         ConfigInt         `yaml:"LootGoblinRoundCount"`         // How often to spawn a loot goblin
	LootGoblinMinimumItems       ConfigInt         `yaml:"LootGoblinMinimumItems"`       // How many items on the ground to attract the loot goblin
	LootGoblinMinimumGold        ConfigInt         `yaml:"LootGoblinMinimumGold"`        // How much gold on the ground to attract the loot goblin
	LootGoblinIncludeRecentRooms ConfigBool        `yaml:"LootGoblinIncludeRecentRooms"` // should the goblin include rooms that have been visited recently?
	LogIntervalRoundCount        ConfigInt         `yaml:"LogIntervalRoundCount"`        // How often to report the current round number.
	Locked                       ConfigSliceString `yaml:"Locked"`                       // List of locked config properties that cannot be changed without editing the file directly.
	Seed                         ConfigString      `yaml:"Seed"`                         // Seed that may be used for generating content
	OnLoginCommands              ConfigSliceString `yaml:"OnLoginCommands"`              // Commands to run when a user logs in
	Motd                         ConfigString      `yaml:"Motd"`                         // Message of the day to display when a user logs in
	BannedNames                  ConfigSliceString `yaml:"BannedNames"`                  // List of names that are not allowed to be used

	TimeFormat ConfigString `yaml:"TimeFormat"` // How to format time when displaying real time

	OnDeathEquipmentDropChance ConfigFloat  `yaml:"OnDeathEquipmentDropChance"` // Chance a player will drop a given piece of equipment on death
	OnDeathAlwaysDropBackpack  ConfigBool   `yaml:"OnDeathAlwaysDropBackpack"`  // If true, players will always drop their backpack items on death
	OnDeathXPPenalty           ConfigString `yaml:"OnDeathXPPenalty"`           // Possible values are: none, level, 10%, 25%, 50%, 75%, 90%, 100%
	EnterRoomMessageWrapper    ConfigString `yaml:"EnterRoomMessageWrapper"`
	ExitRoomMessageWrapper     ConfigString `yaml:"ExitRoomMessageWrapper"`

	MaxIdleSeconds     ConfigInt         `yaml:"MaxIdleSeconds"`     // How many seconds a player can go without a command in game before being kicked.
	TimeoutMods        ConfigBool        `yaml:"TimeoutMods"`        // Whether to kick admin/mods when idle too long.
	ZombieSeconds      ConfigInt         `yaml:"ZombieSeconds"`      // How many seconds a player will be a zombie allowing them to reconnect.
	LogoutRounds       ConfigInt         `yaml:"LogoutRounds"`       // How many rounds of uninterrupted meditation must be completed to log out.
	StartRoom          ConfigInt         `yaml:"StartRoom"`          // Default starting room.
	TutorialStartRooms ConfigSliceString `yaml:"TutorialStartRooms"` // List of all rooms that can be used to begin the tutorial process

	// Perma-death related configs
	PermaDeath     ConfigBool `yaml:"PermaDeath"`     // Is permadeath enabled?
	LivesStart     ConfigInt  `yaml:"LivesStart"`     // Starting permadeath lives
	LivesMax       ConfigInt  `yaml:"LivesMax"`       // Maximum permadeath lives
	LivesOnLevelUp ConfigInt  `yaml:"LivesOnLevelUp"` // # lives gained on level up
	PricePerLife   ConfigInt  `yaml:"PricePerLife"`   // Price in gold to buy new lives

	ShopRestockRate          ConfigString `yaml:"ShopRestockRate"`          // Default time it takes to restock 1 quantity in shops
	ConsistentAttackMessages ConfigBool   `yaml:"ConsistentAttackMessages"` // Whether each weapon has consistent attack messages
	MaxAltCharacters         ConfigInt    `yaml:"MaxAltCharacters"`         // How many characters beyond the default character can they create?
	AfkSeconds               ConfigInt    `yaml:"AfkSeconds"`               // How long until a player is marked as afk?

	LeaderboardSize ConfigInt `yaml:"LeaderboardSize"` // Maximum size of leaderboard

	// Protected values
	turnsPerRound   int     // calculated and cached when data is validated.
	turnsPerSave    int     // calculated and cached when data is validated.
	turnsPerSecond  int     // calculated and cached when data is validated.
	roundsPerMinute float64 // calculated and cached when data is validated.

	overrides map[string]any

	validated bool
}

var (
	configData           Config = Config{overrides: map[string]any{}}
	configDataLock       sync.RWMutex
	ErrInvalidConfigName = errors.New("invalid config name")
	ErrLockedConfig      = errors.New("config name is locked")
)

// Expects a string as the value. Will do the conversion on its own.
func SetVal(propName string, propVal string, force ...bool) error {

	if strings.EqualFold(propName, `locked`) {
		return ErrLockedConfig
	}

	for _, lockedProp := range configData.Locked {
		if strings.EqualFold(lockedProp, propName) {
			if len(force) < 1 || !force[0] {
				return ErrLockedConfig
			}
		}
	}

	overrides := configData.GetOverrides()

	typeSearchStructVal := reflect.ValueOf(configData)
	// Get the value and type of the struct
	//val := reflect.ValueOf(configData)
	typ := typeSearchStructVal.Type()
	// Iterate over all fields of the struct to find the correct name
	for i := 0; i < typeSearchStructVal.NumField(); i++ {
		if strings.EqualFold(typ.Field(i).Name, propName) {
			propName = typ.Field(i).Name
			break
		}
	}

	// Get the reflect.Value of instance
	val := reflect.ValueOf(&configData).Elem() // Use Elem() because we start with a pointer

	// Find the field by name
	fieldVal := val.FieldByName(propName)

	if !fieldVal.IsValid() {
		return fmt.Errorf("no such field: %s in obj", propName)
	}

	// If fieldVal is struct and Set has a pointer receiver, you need to get the address of fieldVal
	if !fieldVal.CanAddr() {
		return fmt.Errorf("field is not addressable")
	}

	fieldValPtr := fieldVal.Addr() // Get a pointer to the field
	method := fieldValPtr.MethodByName("Set")

	if !method.IsValid() {
		return fmt.Errorf("Set method missing")
	}
	// Prepare arguments and call the method as before
	args := []reflect.Value{reflect.ValueOf(propVal)}
	returnValues := method.Call(args)

	// Assuming the method returns an error as its last return value
	if len(returnValues) > 0 { // Check there is at least one return value
		errVal := returnValues[len(returnValues)-1] // Get the last return value
		if errVal.Interface() != nil {              // Check if the returned value is not nil
			if err, ok := errVal.Interface().(error); ok {
				return err
			}
		}
	}

	// set the map value
	reflect.ValueOf(overrides).SetMapIndex(reflect.ValueOf(propName), fieldVal)

	if err := configData.SetOverrides(overrides); err != nil {
		slog.Error("SetVal()", "error", err)
	}

	configData.Validate()

	// save the new config.
	writeBytes, err := yaml.Marshal(configData.GetOverrides())
	if err != nil {
		return err
	}

	overridePath := overridePath()
	return util.Save(overridePath, writeBytes, bool(configData.CarefulSaveFiles))

}

// Get all config data in a map with the field name as the key for easy iteration
func (c Config) AllConfigData(excludeStrings ...string) map[string]any {

	lockedLoookup := map[string]struct{}{
		`locked`: {},
	}
	for _, lockedProp := range configData.Locked {
		lockedLoookup[strings.ToLower(lockedProp)] = struct{}{}
	}

	output := map[string]any{}

	// Get the value and type of the struct
	items := reflect.ValueOf(c)
	typ := items.Type()

	// Iterate over all fields of the struct
	for i := 0; i < items.NumField(); i++ {
		if !items.Field(i).CanInterface() {
			continue
		}

		name := typ.Field(i).Name
		if name == `Locked` {
			continue
		}

		mapName := name

		if _, ok := lockedLoookup[strings.ToLower(name)]; ok {
			mapName = fmt.Sprintf(`%s (locked)`, name)
		}

		if len(excludeStrings) > 0 {
			testName := strings.ToLower(mapName)
			skip := false
			for _, s := range excludeStrings {
				if util.StringWildcardMatch(testName, s) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		itm := items.Field(i)
		if itm.Type().Kind() == reflect.Slice {

			v := reflect.Indirect(itm)
			list := []string{}
			for j := 0; j < v.Len(); j++ {

				cmd := itm.Index(j).Interface().(string)

				if len(excludeStrings) > 0 {

				}
				/*
					if len(cmd) > 27 {
						cmd = cmd[0:27]
					}
				*/
				list = append(list, cmd)
				//output[fmt.Sprintf(`%s.%d`, name, j)] = cmd
			}
			output[name] = strings.Join(list, `; `)

		} else if itm.Type().Kind() == reflect.Map {
			// iterate the map
			keys := itm.MapKeys()
			for _, key := range keys {
				output[fmt.Sprintf(`%s.%d`, name, key.Int())] = itm.MapIndex(key).Float()
			}

		} else {
			output[mapName] = itm.Interface()
		}

	}

	return output
}

func (c *Config) GetOverrides() map[string]any {
	return c.overrides
}

func (c *Config) SetOverrides(overrides map[string]any) error {

	c.overrides = map[string]any{}
	for k, v := range overrides {
		c.overrides[k] = v
	}

	structValue := reflect.ValueOf(c).Elem()
	for name, value := range c.overrides {

		slog.Info("SetOverrides()", "name", name, "value", value)

		structFieldValue := structValue.FieldByName(name)

		if !structFieldValue.IsValid() {
			return fmt.Errorf("No such field: %s in obj", name)
		}

		if !structFieldValue.CanSet() {
			return fmt.Errorf("Cannot set %s field value", name)
		}

		// Get the reflect.Value of instance
		val := reflect.ValueOf(c).Elem() // Use Elem() because we start with a pointer

		// Find the field by name
		fieldVal := val.FieldByName(name)

		if !fieldVal.IsValid() {
			return fmt.Errorf("no such field: %s in obj", name)
		}

		// If fieldVal is struct and Set has a pointer receiver, you need to get the address of fieldVal
		if !fieldVal.CanAddr() {
			return fmt.Errorf("field is not addressable")
		}

		fieldValPtr := fieldVal.Addr() // Get a pointer to the field
		method := fieldValPtr.MethodByName("Set")

		if !method.IsValid() {
			return fmt.Errorf("Set method missing")
		}
		// Prepare arguments and call the method as before
		args := []reflect.Value{reflect.ValueOf(fmt.Sprintf(`%v`, value))}
		method.Call(args)

	}

	return nil
}

// Ensures certain ranges and defaults are observed
func (c *Config) Validate() {

	if c.MaxCPUCores < 0 {
		c.MaxCPUCores = 0 // default
	}

	if c.FolderItemData == `` {
		c.FolderItemData = `_datafiles/items` // default
	}

	if c.FolderAttackMessageData == `` {
		c.FolderAttackMessageData = `_datafiles/combat-messages` // default
	}

	if c.FolderUserData == `` {
		c.FolderUserData = `_datafiles/users` // default
	}

	if c.FolderSpellData == `` {
		c.FolderSpellData = `_datafiles/spells` // default
	}

	if c.FolderTemplates == `` {
		c.FolderTemplates = `_datafiles/templates` // default
	}

	if c.FileAnsiAliases == `` {
		c.FileAnsiAliases = `_datafiles/ansi-aliases.yaml` // default
	}

	if c.FileKeywords == `` {
		c.FileKeywords = `_datafiles/keywords.yaml` // default
	}

	if c.TimeFormat == `` {
		c.TimeFormat = `Monday, 02-Jan-2006 03:04:05PM`
	}

	// Nothing to do with CarefulSaveFiles

	if c.AuctionSeconds < 30 {
		c.AuctionSeconds = 30 // minimum
	}

	if c.AuctionUpdateSeconds < 15 {
		c.AuctionUpdateSeconds = 15 // default
	} else if c.AuctionUpdateSeconds > c.AuctionSeconds>>1 {
		c.AuctionUpdateSeconds = c.AuctionSeconds >> 1 // default
	}

	// Nothing to do with PVPEnabled

	if c.XPScale <= 0 {
		c.XPScale = 1.0 // default
	}

	if c.TurnMs < 10 {
		c.TurnMs = 100 // default
	}

	if c.RoundSeconds < 1 {
		c.RoundSeconds = 4 // default
	}

	if c.OnDeathEquipmentDropChance < 0.0 || c.OnDeathEquipmentDropChance > 1.0 {
		c.OnDeathEquipmentDropChance = 0.0 // default
	}

	// Nothing to do with OnDeathAlwaysDropBackpack

	c.OnDeathXPPenalty.Set(strings.ToLower(string(c.OnDeathXPPenalty)))

	if c.OnDeathXPPenalty != `none` && c.OnDeathXPPenalty != `level` {
		// If not a valid percent, set to default
		if !strings.HasSuffix(string(c.OnDeathXPPenalty), `%`) {
			c.OnDeathXPPenalty = `none` // default
		} else {
			// If not a valid percent, set to default
			percent, err := strconv.ParseInt(string(c.OnDeathXPPenalty)[0:len(c.OnDeathXPPenalty)-1], 10, 64)
			if err != nil || percent < 0 || percent > 100 {
				c.OnDeathXPPenalty = `none` // default
			}
		}
	}

	// Must have a message wrapper...
	if c.EnterRoomMessageWrapper == `` {
		c.EnterRoomMessageWrapper = `%s` // default
	}
	if strings.LastIndex(string(c.EnterRoomMessageWrapper), `%s`) < 0 {
		c.EnterRoomMessageWrapper += `%s` // default
	}

	// Must have a message wrapper...
	if c.ExitRoomMessageWrapper == `` {
		c.ExitRoomMessageWrapper = `%s` // default
	}
	if strings.LastIndex(string(c.ExitRoomMessageWrapper), `%s`) < 0 {
		c.ExitRoomMessageWrapper += `%s` // default
	}

	// Zombie configs
	if c.ZombieSeconds < 0 {
		c.ZombieSeconds = 0 // default
	}
	if c.LogoutRounds < 0 {
		c.LogoutRounds = 3 // default
	}

	if c.RoundsPerAutoSave < 1 {
		c.RoundsPerAutoSave = 900 // default of 15 minutes worth of rounds
	}

	if c.RoundsPerDay < 10 {
		c.RoundsPerDay = 20 // default of 24 hours worth of rounds
	}

	if c.NightHours < 0 {
		c.NightHours = 0
	} else if c.NightHours > 24 {
		c.NightHours = 24
	}

	if c.MaxMobBoredom < 1 {
		c.MaxMobBoredom = 150 // default
	}

	if c.ScriptLoadTimeoutMs < 1 {
		c.ScriptLoadTimeoutMs = 1000 // default
	}

	if c.ScriptRoomTimeoutMs < 1 {
		c.ScriptRoomTimeoutMs = 10
	}

	if c.MaxTelnetConnections < 1 {
		c.MaxTelnetConnections = 50 // default
	}

	if c.WebPort < 1 {
		c.WebPort = 80 // default
	}

	if c.Seed == `` {
		c.Seed = `Mud` // default
	}

	// Nothing to do with NextRoomId

	if c.LootGoblinRoundCount < 10 {
		c.LootGoblinRoundCount = 10 // default
	}

	if c.LootGoblinMinimumItems < 1 {
		c.LootGoblinMinimumItems = 2 // default
	}

	if c.LootGoblinMinimumGold < 1 {
		c.LootGoblinMinimumGold = 100 // default
	}

	// nothing to do with LootGoblinIncludeRecentRooms

	if c.LogIntervalRoundCount < 0 {
		c.LogIntervalRoundCount = 0
	}

	// Nothing to do with Locked

	// Pre-calculate and cache useful values
	c.turnsPerRound = int((c.RoundSeconds * 1000) / c.TurnMs)
	c.turnsPerSave = int(c.RoundsPerAutoSave) * c.turnsPerRound
	c.turnsPerSecond = int(1000 / c.TurnMs)
	c.roundsPerMinute = 60 / float64(c.RoundSeconds)

	c.validated = true
}

func (c Config) GetDeathXPPenalty() (setting string, pct float64) {

	setting = string(c.OnDeathXPPenalty)
	pct = 0.0

	if c.OnDeathXPPenalty == `none` || c.OnDeathXPPenalty == `level` {
		return setting, pct
	}

	percent, err := strconv.ParseInt(string(c.OnDeathXPPenalty)[0:len(c.OnDeathXPPenalty)-1], 10, 64)
	if err != nil || percent < 0 || percent > 100 {
		setting = `none`
		pct = 0.0
		return setting, pct
	}

	pct = float64(percent) / 100.0

	return setting, pct
}

func (c Config) TurnsPerRound() int {
	return c.turnsPerRound
}

func (c Config) TurnsPerAutoSave() int {
	return c.turnsPerSave
}

func (c Config) TurnsPerSecond() int {
	return c.turnsPerSecond
}

func (c Config) MinutesToRounds(minutes int) int {
	return int(math.Ceil(c.roundsPerMinute * float64(minutes)))
}

func (c Config) SecondsToRounds(seconds int) int {
	return int(math.Ceil(float64(seconds) / float64(c.RoundSeconds)))
}

func (c Config) MinutesToTurns(minutes int) int {
	return int(math.Ceil(float64(minutes*60*1000) / float64(c.TurnMs)))
}

func (c Config) SecondsToTurns(seconds int) int {
	return int(math.Ceil(float64(seconds*1000) / float64(c.TurnMs)))
}

func (c Config) RoundsToSeconds(rounds int) int {
	return int(math.Ceil(float64(rounds) * float64(c.RoundSeconds)))
}

func (c Config) IsBannedName(name string) (string, bool) {

	name = strings.ToLower(strings.TrimSpace(name))

	for _, bannedName := range c.BannedNames {
		if util.StringWildcardMatch(name, strings.ToLower(bannedName)) {
			return bannedName, true
		}
	}

	return "", false
}

func GetConfig() Config {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData
}

func overridePath() string {
	overridePath := os.Getenv(`CONFIG_PATH`)
	if overridePath == `` {
		overridePath = `_datafiles/config-overrides.yaml`
	}
	return overridePath
}

func ReloadConfig() error {

	configPath := util.FilePath(defaultConfigPath)

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	tmpConfigData := Config{}
	err = yaml.Unmarshal(bytes, &tmpConfigData)
	if err != nil {
		return err
	}

	overridePath := overridePath()

	slog.Info("ReloadConfig()", "overridePath", overridePath)

	if _, err := os.Stat(util.FilePath(overridePath)); err == nil {
		if overridePath != `` {

			slog.Info("ReloadConfig()", "Loading overrides", true)

			overrideBytes, err := os.ReadFile(util.FilePath(overridePath))
			if err != nil {
				return err
			}

			overrides := map[string]any{}
			err = yaml.Unmarshal(overrideBytes, &overrides)
			if err != nil {
				return err
			}

			if err := tmpConfigData.SetOverrides(overrides); err != nil {
				slog.Error("ReloadConfig()", "error", err)
			}
		}
	} else {
		slog.Info("ReloadConfig()", "Loading overrides", false)
		tmpConfigData.SetOverrides(map[string]any{})
	}

	tmpConfigData.Validate()

	configDataLock.Lock()
	defer configDataLock.Unlock()
	// Assign it
	configData = tmpConfigData

	return nil
}