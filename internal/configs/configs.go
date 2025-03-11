package configs

import (
	"errors"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/util"
	"gopkg.in/yaml.v2"
)

const (
	PVPEnabled  = `enabled`
	PVPDisabled = `disabled`
	PVPOff      = `off`
	PVPLimited  = `limited`
)

var (
	configData Config         = Config{}
	overrides  map[string]any = make(map[string]any)

	configDataLock       sync.RWMutex
	ErrInvalidConfigName = errors.New("invalid config name")
	ErrLockedConfig      = errors.New("config name is locked")
)

type Config struct {
	// Start config subsections
	Server       Server       `yaml:"Server"`
	Memory       Memory       `yaml:"Memory"`
	Auctions     Auctions     `yaml:"Auctions"`
	LootGoblin   LootGoblin   `yaml:"LootGoblin"`
	EngineTiming EngineTiming `yaml:"EngineTiming"`
	FilePaths    FilePaths    `yaml:"FilePaths"`
	GamePlay     GamePlay     `yaml:"GamePlay"`
	Integrations Integrations `yaml:"Integrations"`
	TextFormats  TextFormats  `yaml:"TextFormats"`
	Network      Network      `yaml:"Network"`
	Scripting    Scripting    `yaml:"Scripting"`
	SpecialRooms SpecialRooms `yaml:"SpecialRooms"`
	Statistics   Statistics   `yaml:"Statistics"`
	// End config subsections

	seedInt int64 `yaml:"-"`

	// Protected values
	turnsPerRound   int     // calculated and cached when data is validated.
	turnsPerSave    int     // calculated and cached when data is validated.
	turnsPerSecond  int     // calculated and cached when data is validated.
	roundsPerMinute float64 // calculated and cached when data is validated.

	validated bool
}

// OverlayDotMap overlays values from a dot-syntax map onto the Config.
func (c *Config) OverlayOverrides(dotMap map[string]any) error {
	// First unflatten the dot map into a nested map.
	nestedMap := unflattenMap(dotMap)

	// Marshal the nested map into YAML bytes.
	b, err := yaml.Marshal(nestedMap)
	if err != nil {
		return err
	}

	// Unmarshal the YAML bytes into the existing Config struct.
	return yaml.Unmarshal(b, c)
}

func (c *Config) DotPaths() map[string]interface{} {
	result := make(map[string]interface{})
	// Get the underlying value of c (we assume c is a pointer).
	v := reflect.ValueOf(c).Elem()
	c.buildDotPaths(v, "", result)
	return result
}

func (c *Config) buildDotPaths(v reflect.Value, prefix string, result map[string]interface{}) {
	// If the value is a pointer, dereference it.
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		c.buildDotPaths(v.Elem(), prefix, result)
		return
	}

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			// Skip unexported fields.
			if field.PkgPath != "" {
				continue
			}
			fieldVal := v.Field(i)
			// Determine the key name using the YAML tag if available.
			key := field.Name
			if yamlTag, ok := field.Tag.Lookup("yaml"); ok {
				if tagName := strings.Split(yamlTag, ",")[0]; tagName != "" {
					key = tagName
				}
			}
			// Construct the new prefix.
			newPrefix := key
			if prefix != "" {
				newPrefix = prefix + "." + key
			}
			// Recursively build paths.
			c.buildDotPaths(fieldVal, newPrefix, result)
		}
	default:
		// For non-struct fields, store the value using the accumulated prefix.
		result[prefix] = v.Interface()
	}
}

func GetOverrides() map[string]any {
	return overrides
}

func (c *Config) SetOverrides(newOverrides map[string]any) error {
	overrides = newOverrides
	return nil
}

// Ensures certain ranges and defaults are observed
func (c *Config) Validate() {

	c.Server.Validate()
	c.Memory.Validate()
	c.Auctions.Validate()
	c.LootGoblin.Validate()
	c.EngineTiming.Validate()
	c.FilePaths.Validate()
	c.GamePlay.Validate()
	c.Integrations.Validate()
	c.TextFormats.Validate()
	c.Network.Validate()
	c.Scripting.Validate()
	c.SpecialRooms.Validate()
	c.Statistics.Validate()

	// nothing to do with LootGoblinIncludeRecentRooms

	// Nothing to do with Locked

	// Pre-calculate and cache useful values
	c.turnsPerRound = int((c.EngineTiming.RoundSeconds * 1000) / c.EngineTiming.TurnMs)
	c.turnsPerSave = int(c.EngineTiming.RoundsPerAutoSave) * c.turnsPerRound
	c.turnsPerSecond = int(1000 / c.EngineTiming.TurnMs)
	c.roundsPerMinute = 60 / float64(c.EngineTiming.RoundSeconds)

	c.seedInt = 0
	for i, num := range util.Md5Bytes([]byte(string(c.Server.Seed))) {
		c.seedInt += int64(num) << i
	}

	c.validated = true
}

func (c *Config) setEnvAssignments(clear bool) {

	// We use reflect.Indirect to handle if cfg is a pointer or not
	v := reflect.ValueOf(c).Elem()

	// We'll need the struct type as well (to get field names).
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		if fieldVal.Type().Kind() != reflect.String {
			continue
		}

		if envName := fieldType.Tag.Get(`env`); envName != `` {
			if fieldVal.CanSet() {
				if envVal := os.Getenv(envName); envVal != `` {

					if clear {
						envVal = ``
					}

					fieldVal.Set(reflect.ValueOf(ConfigSecret(envVal)))

				}
			}
		}

	}
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
	return int(math.Ceil(float64(seconds) / float64(c.EngineTiming.RoundSeconds)))
}

func (c Config) MinutesToTurns(minutes int) int {
	return int(math.Ceil(float64(minutes*60*1000) / float64(c.EngineTiming.TurnMs)))
}

func (c Config) SecondsToTurns(seconds int) int {
	return int(math.Ceil(float64(seconds*1000) / float64(c.EngineTiming.TurnMs)))
}

func (c Config) RoundsToSeconds(rounds int) int {
	return int(math.Ceil(float64(rounds) * float64(c.EngineTiming.RoundSeconds)))
}

func (c Config) IsBannedName(name string) (string, bool) {

	name = strings.ToLower(strings.TrimSpace(name))

	for _, bannedName := range c.Server.BannedNames {
		if util.StringWildcardMatch(name, strings.ToLower(bannedName)) {
			return bannedName, true
		}
	}

	return "", false
}

func (c Config) SeedInt() int64 {
	return c.seedInt
}

func (c Config) AllConfigData(excludeStrings ...string) map[string]any {

	finalOutput := make(map[string]any)

	for name, value := range c.DotPaths() {

		if len(excludeStrings) > 0 {
			testName := strings.ToLower(name)
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

		finalOutput[name] = value
	}
	return finalOutput
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
		overridePath = GetConfig().FilePaths.FolderDataFiles.String() + `/config-overrides.yaml`
	}
	return overridePath
}

func ReloadConfig() error {

	configPath := util.FilePath(`_datafiles/config.yaml`)

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

	mudlog.Info("ReloadConfig()", "overridePath", overridePath)

	if _, err := os.Stat(util.FilePath(overridePath)); err == nil {
		if overridePath != `` {

			mudlog.Info("ReloadConfig()", "Loading overrides", true)

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
				mudlog.Error("ReloadConfig()", "error", err)
			}
		}
	} else {
		mudlog.Info("ReloadConfig()", "Loading overrides", false)
		tmpConfigData.SetOverrides(map[string]any{})
	}

	tmpConfigData.setEnvAssignments(false)

	tmpConfigData.Validate()

	configDataLock.Lock()
	defer configDataLock.Unlock()
	// Assign it
	configData = tmpConfigData

	return nil
}

// Usage: configs.GetSecret(c.DiscordWebhookUrl)
func GetSecret(v ConfigSecret) string {
	return string(v)
}
