package configs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TestOverlay tests overlaying a nested map into the Config struct.
func TestOverlay(t *testing.T) {
	// Start with a default config.
	cfg := Config{
		Validation: Validation{
			NameRejectRegex: "test",
		},
	}

	newValues := map[string]any{
		"Validation": map[string]any{
			"NameRejectRegex": "test-changed",
		},
	}

	if err := cfg.OverlayOverrides(newValues); err != nil {
		t.Fatalf("Overlay failed: %v", err)
	}

	if cfg.Validation.NameRejectRegex != "test-changed" {
		t.Errorf("Expected NameRejectRegex to be \"test-changed\", got \"%s\"", cfg.Validation.NameRejectRegex)
	}
}

// TestOverlayDotMap tests overlaying a configuration using dot-syntax keys.
func TestOverlayDotMap(t *testing.T) {
	// Start with a default config.
	cfg := Config{
		Validation: Validation{
			NameRejectRegex: "test",
		},
	}

	dotValues := map[string]any{
		"Validation.NameRejectRegex": "test-changed",
	}

	if err := cfg.OverlayOverrides(dotValues); err != nil {
		t.Fatalf("OverlayDotMap failed: %v", err)
	}

	if cfg.Validation.NameRejectRegex != "test-changed" {
		t.Errorf("Expected LeaderboardSize to be \"test-changed\", got \"%s\"", cfg.Validation.NameRejectRegex)
	}
}

// TestOverlayDotMapMultipleFields demonstrates overlaying multiple fields using dot-syntax.
// Here, we extend the configuration to have an additional field.
func TestOverlayDotMapMultipleFields(t *testing.T) {
	// Define an extended configuration.
	type ExtendedStatistics struct {
		LeaderboardSize int    `yaml:"LeaderboardSize"`
		SomeField       string `yaml:"SomeField"`
	}

	type ExtendedConfig struct {
		Statistics ExtendedStatistics `yaml:"Statistics"`
	}

	cfg := ExtendedConfig{
		Statistics: ExtendedStatistics{
			LeaderboardSize: 5,
			SomeField:       "default",
		},
	}

	dotValues := map[string]any{
		"Statistics.LeaderboardSize": 25,
		"Statistics.SomeField":       "updated",
	}

	// Unflatten the dot-syntax map.
	nestedMap := unflattenMap(dotValues)
	// Marshal to YAML and then unmarshal into the extended config.
	b, err := yaml.Marshal(nestedMap)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if cfg.Statistics.LeaderboardSize != 25 {
		t.Errorf("Expected LeaderboardSize to be 25, got %d", cfg.Statistics.LeaderboardSize)
	}
	if cfg.Statistics.SomeField != "updated" {
		t.Errorf("Expected SomeField to be 'updated', got '%s'", cfg.Statistics.SomeField)
	}
}

// TestAddOverlayOverridesPreservesOperatorOverrides reproduces the boot
// sequence for a world whose config-overrides.yaml contains a partial
// Modules block: the operator has set some (but not all) of a module's keys,
// and the module's data-overlay config ships a superset of those keys.
// Applying the overlay must fill in the missing defaults WITHOUT clobbering
// the operator-supplied values or other modules' blocks.
func TestAddOverlayOverridesPreservesOperatorOverrides(t *testing.T) {

	// Snapshot and restore package globals so this test is hermetic.
	origConfigData := configData
	origOverrides := overrides
	origKeyLookups := keyLookups
	origTypeLookups := typeLookups
	t.Cleanup(func() {
		configData = origConfigData
		overrides = origOverrides
		keyLookups = origKeyLookups
		typeLookups = origTypeLookups
	})

	configData = Config{}
	keyLookups = map[string]string{}
	typeLookups = map[string]string{}

	// The operator's config-overrides.yaml: a partial Modules block for
	// "weather" (missing NewSetting), plus an unrelated module's block.
	operatorYAML := []byte(`
Modules:
  weather:
    Enabled: true
    CycleSeconds: 120
  othermod:
    Setting: keepme
`)
	loadedOverrides := map[string]any{}
	require.NoError(t, yaml.Unmarshal(operatorYAML, &loadedOverrides))
	overrides = loadedOverrides

	// ReloadConfig applies operator overrides onto the live config at boot.
	require.NoError(t, configData.OverlayOverrides(overrides))

	// The module loads and registers its data-overlay config, a superset of
	// the operator's keys. This mirrors how internal/plugins builds the map.
	err := AddOverlayOverrides(map[string]any{
		`Modules.weather.Enabled`:      false,        // module default; operator set true
		`Modules.weather.CycleSeconds`: 60,           // module default; operator set 120
		`Modules.weather.NewSetting`:   `overlayval`, // new key, absent from operator overrides
	})
	require.NoError(t, err)

	flat := Flatten(map[string]any(configData.Modules))

	// (a) Operator-supplied values must survive in the live config.
	require.Equal(t, true, flat[`weather.Enabled`], `operator override Modules.weather.Enabled was clobbered by the module overlay`)
	require.Equal(t, 120, flat[`weather.CycleSeconds`], `operator override Modules.weather.CycleSeconds was clobbered by the module overlay`)

	// (b) Keys absent from the operator overrides get the overlay defaults.
	require.Equal(t, `overlayval`, flat[`weather.NewSetting`])

	// (c) Other modules' blocks are untouched.
	require.Equal(t, `keepme`, flat[`othermod.Setting`])
}
