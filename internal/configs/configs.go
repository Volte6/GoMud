package configs

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/util"
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

	keyLookups  map[string]string = map[string]string{}
	typeLookups map[string]string = map[string]string{}

	configDataLock       sync.RWMutex
	ErrInvalidConfigName = errors.New("invalid config name")
	ErrLockedConfig      = errors.New("config name is locked")
	ErrRedactedValue     = errors.New("cannot save a redacted placeholder as a config value")

	onChangedCallbacks []func(key string)
)

type Config struct {
	// Start config subsections
	Server       Server       `yaml:"Server"`
	Memory       Memory       `yaml:"Memory"`
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
	Validation   Validation   `yaml:"Validation"`
	// Plugins is a special case
	Modules Modules `yaml:"Modules"`

	// End config subsections

	seedInt int64 `yaml:"-"`

	validated bool
}

func AddOverlayOverrides(dotMap map[string]any) error {
	configDataLock.Lock()
	defer configDataLock.Unlock()

	// Only register type/key lookups for every key in the overlay, but only
	// write values for keys that have not already been set by a user override
	// (e.g. config-overrides.yaml). This ensures module defaults never clobber
	// operator-supplied values.
	newKeys := map[string]any{}
	flatOverrides := Flatten(overrides)
	for k, v := range dotMap {
		addKeyLookups(keyLookups, k)
		typeLookups[k] = reflect.TypeOf(v).String()
		if _, alreadySet := flatOverrides[k]; !alreadySet {
			overrides[k] = v
			newKeys[k] = v
		}
	}

	if len(newKeys) == 0 {
		return nil
	}

	// Overlay the full override set rather than just the new keys. Overlaying
	// only the new keys would unmarshal a partial Modules block into the live
	// config, replacing the inner Modules[<name>] map wholesale and discarding
	// every operator-supplied value for that module. Flatten first: at this
	// point `overrides` is mixed-shape (nested maps loaded from
	// config-overrides.yaml plus the flat dotted keys added above), and
	// OverlayOverrides re-nests it internally.
	return configData.OverlayOverrides(Flatten(overrides))
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

func (c *Config) DotPaths() map[string]any {
	result := make(map[string]any)
	// Get the underlying value of c (we assume c is a pointer).
	v := reflect.ValueOf(c).Elem()
	c.buildDotPaths(v, "", result)
	return result
}

func (c *Config) buildDotPaths(v reflect.Value, prefix string, result map[string]any) {
	// If the value is a pointer, dereference it.
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
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
	case reflect.Map:
		// If the map is nil, store nil for the current prefix.
		if v.IsNil() {
			result[prefix] = make(map[string]any)
			return
		}
		// Iterate over each key in the map.
		for _, key := range v.MapKeys() {
			// Convert the key to string (works for string keys; for others, fmt.Sprintf is used).
			keyStr := fmt.Sprintf("%v", key.Interface())
			newPrefix := keyStr
			if prefix != "" {
				newPrefix = prefix + "." + keyStr
			}
			mapVal := v.MapIndex(key)
			// Unwrap interface values so the underlying kind is visible.
			if mapVal.Kind() == reflect.Interface {
				mapVal = mapVal.Elem()
			}
			// YAML unmarshaling produces map[interface{}]interface{} for nested
			// maps. Convert those to map[string]any so they are JSON-encodable.
			if mapVal.IsValid() && mapVal.Kind() == reflect.Map && mapVal.Type().Key().Kind() == reflect.Interface {
				converted := make(map[string]any)
				for _, mk := range mapVal.MapKeys() {
					converted[fmt.Sprintf("%v", mk.Interface())] = mapVal.MapIndex(mk).Interface()
				}
				mapVal = reflect.ValueOf(converted)
			}
			c.buildDotPaths(mapVal, newPrefix, result)
		}
	default:
		// For non-struct fields, store the value using the accumulated prefix.
		// Preserve typed config values so that typeLookups records the correct
		// type name (e.g. "configs.ConfigSliceString") and SetVal can round-trip
		// them. Only sanitize values that are not already a known config type.
		if cs, ok := v.Interface().(ConfigSliceString); ok {
			result[prefix] = cs
		} else {
			// Sanitize to ensure all values are JSON-encodable (YAML unmarshaling
			// can produce map[interface{}]interface{} inside slices and nested maps).
			result[prefix] = sanitizeForJSON(v.Interface())
		}
	}
}

func GetOverrides() map[string]any {
	configDataLock.RLock()
	defer configDataLock.RUnlock()
	return overrides
}

// RestoreOverrides replaces the current overrides with the supplied flat
// dot-path map and re-applies them to the live config. It is used by the
// test-mode middleware to revert changes made during a dry-run request.
func RestoreOverrides(flatSnapshot map[string]any) error {
	configDataLock.Lock()
	defer configDataLock.Unlock()

	overrides = unflattenMap(flatSnapshot)
	if err := configData.OverlayOverrides(overrides); err != nil {
		return err
	}
	configData.Validate()
	return nil
}

func (c *Config) SetOverrides(newOverrides map[string]any) error {

	overrides = newOverrides
	return c.OverlayOverrides(overrides)
}

// Ensures certain ranges and defaults are observed
func (c *Config) Validate() {

	c.Server.Validate()
	c.Memory.Validate()
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
	c.Validation.Validate()
	c.Modules.Validate()

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

// OnChanged registers a callback that is invoked after SetVal successfully
// applies a config change. The callback receives the full dot-path key.
func OnChanged(fn func(key string)) {
	onChangedCallbacks = append(onChangedCallbacks, fn)
}

func SetVal(propertyPath string, newVal string) error {

	if newVal == RedactedValue {
		return ErrRedactedValue
	}

	configDataLock.Lock()
	defer configDataLock.Unlock()

	propertyPath, propertyType := findFullPathFrom(propertyPath, keyLookups, typeLookups)
	if propertyType == `` {
		return errors.New(`invalid property name: ` + propertyPath)
	}

	quickMap := make(map[string]any)
	quickMap[propertyPath] = StringToConfigValue(newVal, propertyType)

	flatOverrides := Flatten(overrides)
	flatQuickmap := Flatten(quickMap)

	for k, v := range flatQuickmap {
		flatOverrides[k] = v
	}

	overrides = unflattenMap(flatOverrides)

	writeBytes, err := yaml.Marshal(overrides)
	if err != nil {
		return err
	}

	ovrPath := overridePathNoLock()
	if err := util.Save(ovrPath, writeBytes, bool(configData.FilePaths.CarefulSaveFiles)); err != nil {
		return err
	}

	if err = configData.OverlayOverrides(overrides); err != nil {
		return err
	}

	configData.Validate()

	for _, fn := range onChangedCallbacks {
		fn(propertyPath)
	}

	return nil
}

func GetConfig() Config {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData
}

// Caller must hold at least configDataLock.RLock.
func overridePathNoLock() string {
	p := os.Getenv(`CONFIG_PATH`)
	if p == `` {
		p = configData.FilePaths.DataFiles.String() + `/config-overrides.yaml`
	}
	return p
}

func ReloadConfig() error {

	configPath := `_datafiles/config.yaml`

	bytes, err := util.ReadFile(configPath)
	if err != nil {
		return err
	}

	tmpConfigData := Config{}
	err = yaml.Unmarshal(bytes, &tmpConfigData)
	if err != nil {
		return err
	}

	// Snapshot current config under read lock for building lookups
	configDataLock.RLock()
	allData := configData.AllConfigData()
	ovrPath := overridePathNoLock()
	configDataLock.RUnlock()

	tmpKeyLookups := make(map[string]string, len(allData)*3)
	tmpTypeLookups := make(map[string]string, len(allData))
	for k, v := range allData {
		addKeyLookups(tmpKeyLookups, k)
		tmpTypeLookups[k] = reflect.TypeOf(v).String()
	}

	mudlog.Info("ReloadConfig()", "overridePath", ovrPath)

	var tmpOverrides map[string]any
	if _, err := os.Stat(util.FilePath(ovrPath)); err == nil {
		if ovrPath != `` {

			mudlog.Info("ReloadConfig()", "Loading overrides", true)

			overrideBytes, err := util.ReadFile(ovrPath)
			if err != nil {
				return err
			}

			tmpOverrides = map[string]any{}
			err = yaml.Unmarshal(overrideBytes, &tmpOverrides)
			if err != nil {
				return err
			}

			for k, v := range tmpOverrides {
				if newKey, _ := findFullPathFrom(k, tmpKeyLookups, tmpTypeLookups); newKey != k {
					tmpOverrides[newKey] = v
					delete(tmpOverrides, k)
				}
			}

			tmpConfigData.OverlayOverrides(tmpOverrides)
		}
	} else {
		mudlog.Info("ReloadConfig()", "Loading overrides", false)
	}

	tmpConfigData.setEnvAssignments(false)

	tmpConfigData.Validate()

	configDataLock.Lock()
	defer configDataLock.Unlock()
	configData = tmpConfigData
	keyLookups = tmpKeyLookups
	typeLookups = tmpTypeLookups
	if tmpOverrides != nil {
		overrides = tmpOverrides
	}

	return nil
}

func addKeyLookups(keys map[string]string, k string) {
	if strings.Contains(k, `.`) {
		parts := strings.Split(k, `.`)
		for i := len(parts) - 1; i >= 0; i-- {
			dotSuffix := strings.Join(parts[i:], `.`)
			keys[strings.ToLower(dotSuffix)] = k
			noDotSuffix := strings.Join(parts[i:], ``)
			keys[strings.ToLower(noDotSuffix)] = k
		}
	} else {
		keys[strings.ToLower(k)] = k
	}
}

func findFullPathFrom(inputKey string, keys map[string]string, types map[string]string) (string, string) {
	if v, ok := keys[strings.ToLower(inputKey)]; ok {
		return v, types[v]
	}
	return inputKey, types[inputKey]
}

func FindFullPath(inputKey string) (properKey string, typeName string) {
	return findFullPathFrom(inputKey, keyLookups, typeLookups)
}

// Usage: configs.GetSecret(c.DiscordWebhookUrl)
func GetSecret(v ConfigSecret) string {
	return string(v)
}

// sanitizeForJSON recursively converts map[interface{}]interface{} (produced
// by YAML unmarshaling) to map[string]any so that values stored in AllConfigData
// are always JSON-encodable.
func sanitizeForJSON(v any) any {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]any, len(val))
		for k, mv := range val {
			out[fmt.Sprintf("%v", k)] = sanitizeForJSON(mv)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, mv := range val {
			out[k] = sanitizeForJSON(mv)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, elem := range val {
			out[i] = sanitizeForJSON(elem)
		}
		return out
	}
	return v
}

// flatten recursively flattens a map[string]any.
// It supports both map[string]any and map[any]any values,
// which is useful when unmarshaling YAML.
func Flatten(input map[string]any) map[string]any {
	flatMap := make(map[string]any)
	flattenHelper("", input, flatMap)
	return flatMap
}

// flattenHelper is a recursive helper that constructs the flattened map.
func flattenHelper(prefix string, input map[string]any, flatMap map[string]any) {
	for key, value := range input {
		// Construct the new key path.
		var newKey string
		if prefix == "" {
			newKey = key
		} else {
			newKey = prefix + "." + key
		}

		// Handle nested maps from YAML unmarshaling, which can be of type map[any]any.
		switch v := value.(type) {
		case map[string]any:
			flattenHelper(newKey, v, flatMap)
		case map[any]any:
			// Convert map[any]any to map[string]any.
			converted := make(map[string]any)
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
