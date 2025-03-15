package configs

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/util"
	"gopkg.in/yaml.v3"
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

	keyLookups  map[string]string = map[string]string{}
	typeLookups map[string]string = map[string]string{}

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
	Timing       Timing       `yaml:"Timing"`
	FilePaths    FilePaths    `yaml:"FilePaths"`
	GamePlay     GamePlay     `yaml:"GamePlay"`
	Integrations Integrations `yaml:"Integrations"`
	TextFormats  TextFormats  `yaml:"TextFormats"`
	Translation  Translation  `yaml:"Translation"`
	Network      Network      `yaml:"Network"`
	Scripting    Scripting    `yaml:"Scripting"`
	SpecialRooms SpecialRooms `yaml:"SpecialRooms"`
	Statistics   Statistics   `yaml:"Statistics"`
	Validation   Validation   `yaml:"Validation"`

	// End config subsections

	seedInt int64 `yaml:"-"`

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
	c.OverlayOverrides(overrides)

	return nil
}

// Ensures certain ranges and defaults are observed
func (c *Config) Validate() {

	c.Server.Validate()
	c.Memory.Validate()
	c.Auctions.Validate()
	c.LootGoblin.Validate()
	c.Timing.Validate()
	c.FilePaths.Validate()
	c.GamePlay.Validate()
	c.Integrations.Validate()
	c.TextFormats.Validate()
	c.Translation.Validate()
	c.Network.Validate()
	c.Scripting.Validate()
	c.SpecialRooms.Validate()
	c.Statistics.Validate()
	c.Validation.Validate()

	// nothing to do with LootGoblinIncludeRecentRooms

	// Nothing to do with Locked

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

func (c Config) IsBannedName(name string) (string, bool) {

	name = strings.ToLower(strings.TrimSpace(name))

	for _, bannedName := range c.Validation.BannedNames {
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

func SetVal(propertyPath string, newVal string) error {

	propertyPath, propertyType := FindFullPath(propertyPath)

	quickMap := make(map[string]any)
	quickMap[propertyPath] = StringToConfigValue(newVal, propertyType)

	flatOverrides := flatten(overrides)
	flatQuickmap := flatten(quickMap)

	for k, v := range flatQuickmap {
		flatOverrides[k] = v
	}

	overrides = unflattenMap(flatOverrides)

	// save the new config.
	writeBytes, err := yaml.Marshal(overrides)
	if err != nil {
		return err
	}

	overridePath := overridePath()
	if err := util.Save(overridePath, writeBytes, bool(configData.FilePaths.CarefulSaveFiles)); err != nil {
		return err
	}

	return configData.OverlayOverrides(overrides)
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

	// Build a special lookup to attempt to match old data or even some minor typos
	keyLookups = map[string]string{}
	typeLookups = map[string]string{}
	for k, v := range configData.AllConfigData() {

		if strings.Index(k, `.`) != -1 {

			parts := strings.Split(k, `.`)

			for i := len(parts) - 1; i >= 0; i-- {
				tmpKey := strings.Join(parts[i:], `.`)
				keyLookups[strings.ToLower(tmpKey)] = k

				tmpKey = strings.Join(parts[i:], ``)
				keyLookups[strings.ToLower(tmpKey)] = k

			}

		} else {
			keyLookups[strings.ToLower(k)] = k
		}

		typeLookups[k] = reflect.TypeOf(v).String()
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

			tmpOverrides := map[string]any{}
			err = yaml.Unmarshal(overrideBytes, &tmpOverrides)
			if err != nil {
				return err
			}

			// Attempt a correction for bad names
			for k, v := range tmpOverrides {
				if newKey, _ := FindFullPath(k); newKey != k {
					tmpOverrides[newKey] = v
					delete(tmpOverrides, k)
				}
			}

			tmpConfigData.SetOverrides(tmpOverrides)
		}
	} else {
		mudlog.Info("ReloadConfig()", "Loading overrides", false)
	}

	tmpConfigData.setEnvAssignments(false)

	tmpConfigData.Validate()

	configDataLock.Lock()
	defer configDataLock.Unlock()
	// Assign it
	configData = tmpConfigData

	return nil
}

func FindFullPath(inputKey string) (properKey string, typeName string) {

	if v, ok := keyLookups[strings.ToLower(inputKey)]; ok {
		return v, typeLookups[v]
	}
	return inputKey, typeLookups[inputKey]
}

// Usage: configs.GetSecret(c.DiscordWebhookUrl)
func GetSecret(v ConfigSecret) string {
	return string(v)
}

// flatten recursively flattens a map[string]interface{}.
// It supports both map[string]interface{} and map[interface{}]interface{} values,
// which is useful when unmarshaling YAML.
func flatten(input map[string]interface{}) map[string]interface{} {
	flatMap := make(map[string]interface{})
	flattenHelper("", input, flatMap)
	return flatMap
}

// flattenHelper is a recursive helper that constructs the flattened map.
func flattenHelper(prefix string, input map[string]interface{}, flatMap map[string]interface{}) {
	for key, value := range input {
		// Construct the new key path.
		var newKey string
		if prefix == "" {
			newKey = key
		} else {
			newKey = prefix + "." + key
		}

		// Handle nested maps from YAML unmarshaling, which can be of type map[interface{}]interface{}.
		switch v := value.(type) {
		case map[string]interface{}:
			flattenHelper(newKey, v, flatMap)
		case map[interface{}]interface{}:
			// Convert map[interface{}]interface{} to map[string]interface{}.
			converted := make(map[string]interface{})
			for k, val := range v {
				if strKey, ok := k.(string); ok {
					converted[strKey] = val
				}
			}
			flattenHelper(newKey, converted, flatMap)
		default:
			flatMap[newKey] = value
		}
	}
}
