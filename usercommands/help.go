package usercommands

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/volte6/mud/keywords"
	"github.com/volte6/mud/races"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Help(rest string, userId int) (bool, error) {

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return false, fmt.Errorf(`user %d not found`, userId)
	}

	var helpTxt string
	var err error = nil

	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {

		type helpCommand struct {
			Command string
			Type    string
			Missing bool
		}

		type commandLists struct {
			Commands map[string][]helpCommand
			Skills   map[string][]helpCommand
			Admin    map[string][]helpCommand
		}

		helpCommandList := commandLists{
			Commands: make(map[string][]helpCommand),
			Skills:   make(map[string][]helpCommand),
			Admin:    make(map[string][]helpCommand),
		}

		for _, command := range keywords.GetAllHelpTopicInfo() {

			category := command.Category
			if category == `all` {
				category = ``
			}

			templateFile := `help/` + keywords.TryHelpAlias(command.Command)

			if command.AdminOnly {
				if user.Permission == users.PermissionAdmin || user.HasAdminCommand(command.Command) {
					helpCommandList.Admin[category] = append(
						helpCommandList.Admin[category],
						helpCommand{Command: command.Command, Type: "command-admin", Missing: !templates.Exists(templateFile)},
					)
				}
				continue
			}

			hlpCmd := helpCommand{Command: command.Command, Type: command.Type, Missing: !templates.Exists(templateFile)}

			if command.Type == `skill` {
				helpCommandList.Skills[category] = append(helpCommandList.Skills[category], hlpCmd)
				continue
			}

			helpCommandList.Commands[category] = append(helpCommandList.Commands[category], hlpCmd)

		}

		helpTxt, err = templates.Process("help/help", helpCommandList)
		if err != nil {
			helpTxt = err.Error()
		}
	} else {

		helpName := args[0]
		helpRest := ``

		args := args[1:]
		if len(args) > 0 {
			helpRest = strings.Join(args, ` `)
		}

		// replace any non alpha/numeric characters in "rest"
		helpName = regexp.MustCompile(`[^a-zA-Z0-9\\-]+`).ReplaceAllString(helpName, ``)

		helpName = keywords.TryHelpAlias(helpName)

		var helpVars any = nil

		if helpName == `emote` {
			helpVars = emoteAliases
		}

		if helpName == `races` {
			helpVars = getRaceOptions(helpRest)
		}

		if helpName == `spell` {
			sData := spells.GetSpell(helpRest)
			if sData == nil {
				sData = spells.FindSpellByName(helpRest)
			}

			if sData == nil {
				helpName = `spells`
			} else {
				helpVars = *sData
			}
		}

		helpTxt, err = templates.Process("help/"+helpName, helpVars)
		if err != nil {
			user.SendText(fmt.Sprintf(`No help found for "%s"`, helpName))
			return true, err
		}
	}

	user.SendText(helpTxt)

	return true, nil
}

func getRaceOptions(raceRequest string) []races.Race {

	allRaces := races.GetRaces()
	sort.Slice(allRaces, func(i, j int) bool {
		return allRaces[i].RaceId < allRaces[j].RaceId
	})

	raceNames := strings.Split(raceRequest, ` `)

	getAllRaces := false
	if raceRequest == `all` {
		getAllRaces = true
	}

	raceOptions := []races.Race{}
	for _, race := range allRaces {

		if len(raceRequest) == 0 {
			if !race.Selectable && !getAllRaces {
				continue
			}
		} else if len(raceNames) > 0 {
			lowerName := strings.ToLower(race.Name)
			found := false
			for _, rName := range raceNames {
				if strings.Contains(lowerName, strings.ToLower(rName)) {
					found = true
					break
				}
			}
			if !getAllRaces && !found {
				continue
			}
		}
		raceOptions = append(raceOptions, race)
	}

	return raceOptions
}
