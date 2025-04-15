package audio

import (
	"os"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type AudioConfig struct {
	FilePath string `yaml:"filepath,omitempty"`
	Volume   int    `yaml:"volume,omitempty"`
}

var (
	audioLookup = map[string]AudioConfig{}
)

func GetFile(identifier string) AudioConfig {
	if f, ok := audioLookup[identifier]; ok {
		return f
	}
	return AudioConfig{}
}

func LoadAudioConfig() {

	start := time.Now()

	path := string(configs.GetFilePathsConfig().DataFiles) + `/audio.yaml`

	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(errors.Wrap(err, `filepath: `+path))
	}

	clear(audioLookup)

	err = yaml.Unmarshal(bytes, &audioLookup)
	if err != nil {
		panic(errors.Wrap(err, `filepath: `+path))
	}

	mudlog.Info("...LoadAudioConfig()", "loadedCount", len(audioLookup), "Time Taken", time.Since(start))
}
