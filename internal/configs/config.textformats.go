package configs

import (
	"strings"
)

type TextFormats struct {
	Prompt                  ConfigString      `yaml:"Prompt"`                  // The in-game status prompt style
	EnterRoomMessageWrapper ConfigString      `yaml:"EnterRoomMessageWrapper"` // Special enter messages
	ExitRoomMessageWrapper  ConfigString      `yaml:"ExitRoomMessageWrapper"`  // Special exit messages
	Time                    ConfigString      `yaml:"Time"`                    // How to format time when displaying real time
	TimeShort               ConfigString      `yaml:"TimeShort"`               // How to format time when displaying real time (shortform)
	DefaultLanguage         ConfigString      `yaml:"DefaultLanguage"`         // Specify the default game language (fallback)
	Language                ConfigString      `yaml:"Language"`                // Specify the game language
	LanguagePaths           ConfigSliceString `yaml:"LanguagePaths"`           // Specify the game language file paths
}

func (m *TextFormats) Validate() {

	if m.Prompt == `` {
		m.Prompt = `{8}[{t} {T} {255}HP:{hp}{8}/{HP} {255}MP:{13}{mp}{8}/{13}{MP}{8}]{239}{h}{8}:`
	}

	// Must have a message wrapper...
	if m.EnterRoomMessageWrapper == `` {
		m.EnterRoomMessageWrapper = `%s` // default
	}
	if strings.LastIndex(string(m.EnterRoomMessageWrapper), `%s`) < 0 {
		m.EnterRoomMessageWrapper += `%s` // default
	}

	// Must have a message wrapper...
	if m.ExitRoomMessageWrapper == `` {
		m.ExitRoomMessageWrapper = `%s` // default
	}
	if strings.LastIndex(string(m.ExitRoomMessageWrapper), `%s`) < 0 {
		m.ExitRoomMessageWrapper += `%s` // default
	}

	if m.Time == `` {
		m.Time = `Monday, 02-Jan-2006 03:04:05PM`
	}

	if m.TimeShort == `` {
		m.TimeShort = `Jan 2 '06 3:04PM`
	}

	if m.DefaultLanguage == `` {
		m.Language = `en` // default
	}

	if m.Language == `` {
		m.Language = `en` // default
	}

}

func GetTextFormatsConfig() TextFormats {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.TextFormats
}
