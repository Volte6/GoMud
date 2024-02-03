package keywords

import (
	"sort"
	"strings"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/fileloader"
)

type HelpTopic struct {
	Command   string
	Type      string // command/skill
	Category  string
	AdminOnly bool
}

type Aliases struct {
	Help             map[string]map[string][]string `yaml:"help"`
	HelpAliases      map[string][]string            `yaml:"help-aliases"`
	CommandAliases   map[string][]string            `yaml:"command-aliases"`
	DirectionAliases map[string]string              `yaml:"direction-aliases"`

	// helpTopics
	// [skill/command/admin]
	// helpTopics[`skill`][`character`][`alignment`] = `alignment`
	helpTopics map[string]HelpTopic
	// Organized help aliases
	helpAliases map[string]string
	// Organized command aliases
	commandAliases map[string]string
}

// Presumably to ensure the datafile hasn't messed something up.
func (a *Aliases) Validate() error {

	//
	// Unroll the data into structures that are quickly searched
	//

	a.helpTopics = map[string]HelpTopic{}

	// helpGroup = commands/skills/admin
	for helpGroup, helpTypes := range a.Help {

		helpGroup = strings.ToLower(helpGroup)

		// helpType = configuration/character/shops/quests/combat
		for helpCategory, helpList := range helpTypes {
			for _, helpCommand := range helpList {

				helpCommand = strings.Title(strings.ToLower(helpCommand))

				entry := HelpTopic{
					Command:   helpCommand,
					Type:      helpGroup,
					Category:  helpCategory,
					AdminOnly: (helpGroup == `admin`),
				}

				a.helpTopics[helpCommand] = entry
			}
		}
	}

	a.helpAliases = map[string]string{}
	for helpTopic, aliasList := range a.HelpAliases {
		helpTopic = strings.ToLower(helpTopic)
		for _, alias := range aliasList {
			a.helpAliases[alias] = helpTopic
		}
	}

	a.commandAliases = map[string]string{}
	for command, aliasList := range a.CommandAliases {
		command = strings.ToLower(command)
		for _, alias := range aliasList {
			a.commandAliases[alias] = command
		}
	}

	//for alias, direction := range a.DirectionAliases {
	//	a.commandAliases[strings.ToLower(alias)] = direction
	//}

	return nil
}

func (a *Aliases) Filename() string {
	return `keywords.yaml`
}

func (a *Aliases) Filepath() string {
	return a.Filename()
}

func GetAllHelpTopics() []string {
	helpTopics := []string{}

	for helpCommand, _ := range loadedKeywords.helpTopics {
		helpTopics = append(helpTopics, helpCommand)
	}

	sort.Slice(helpTopics, func(i, j int) bool {
		return helpTopics[i] < helpTopics[j]
	})

	return helpTopics
}

func GetAllHelpTopicInfo() []HelpTopic {
	helpTopics := []HelpTopic{}
	for _, helpTopicInfo := range loadedKeywords.helpTopics {
		helpTopics = append(helpTopics, helpTopicInfo)
	}

	sort.Slice(helpTopics, func(i, j int) bool {
		return helpTopics[i].Command < helpTopics[j].Command
	})

	return helpTopics
}

func GetAllCommandAliases() map[string]string {

	ret := map[string]string{}
	for alias, command := range loadedKeywords.commandAliases {
		ret[alias] = command
	}

	return ret
}

func GetAllHelpAliases() map[string]string {

	ret := map[string]string{}
	for alias, command := range loadedKeywords.helpAliases {
		ret[alias] = command
	}

	return ret
}

func TryDirectionAlias(input string) string {
	if alias, ok := loadedKeywords.DirectionAliases[strings.ToLower(input)]; ok {
		return alias
	}

	return input
}

func TryCommandAlias(input string) string {
	if alias, ok := loadedKeywords.commandAliases[strings.ToLower(input)]; ok {
		return alias
	}

	return input
}

func TryHelpAlias(input string) string {
	if alias, ok := loadedKeywords.helpAliases[strings.ToLower(input)]; ok {
		return alias
	}
	return input
}

var (
	loadedKeywords *Aliases
)

// Loads the ansi aliases from the config file
// Only if the file has been modified since the last load
func LoadAliases() {

	var err error
	loadedKeywords, err = fileloader.LoadFlatFile[*Aliases](configs.GetConfig().FileKeywords)
	if err != nil {
		panic(err)
	}

}
