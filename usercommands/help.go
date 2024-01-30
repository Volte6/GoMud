package usercommands

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/volte6/mud/races"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Help(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
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

		for i := 0; i < len(helpCommands); i++ {

			category := helpCommands[i].Category

			command := helpCommands[i]
			templateFile := `help/` + command.Command
			if newHelp, ok := helpAliases[command.Command]; ok {
				templateFile = `help/` + newHelp
			}

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

			if command.Type == "skill" {
				helpCommandList.Skills[category] = append(helpCommandList.Skills[category], hlpCmd)
				continue
			}

			helpCommandList.Commands[category] = append(helpCommandList.Commands[category], hlpCmd)
		}

		helpTxt, err = templates.Process("help/help", helpCommandList)
		if err != nil {
			response.SendUserMessage(userId, err.Error(), true)
			response.Handled = true
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

		if newHelp, ok := helpAliases[helpName]; ok {
			helpName = newHelp
		}

		var helpVars any = nil

		if helpName == `emote` {
			helpVars = emoteAliases
		}

		if helpName == `races` || helpName == `setrace` {
			helpVars = getRaceOptions(helpRest)
		}

		helpTxt, err = templates.Process("help/"+helpName, helpVars)
		if err != nil {
			response.SendUserMessage(userId, fmt.Sprintf(`No help found for "%s"`, helpName), true)
			response.Handled = true
			return response, err
		}
	}

	response.SendUserMessage(userId, helpTxt, false)

	response.Handled = true
	return response, nil
}

func getRaceOptions(raceRequest string) []races.Race {

	allRaces := races.GetRaces()
	sort.Slice(allRaces, func(i, j int) bool {
		return allRaces[i].RaceId < allRaces[j].RaceId
	})

	raceOptions := []races.Race{}
	for _, race := range allRaces {

		if len(raceRequest) == 0 {
			if !race.Selectable && raceRequest != `all` {
				continue
			}
		} else if len(raceRequest) > 0 {
			lowerName := strings.ToLower(race.Name)
			if raceRequest != `all` && !strings.Contains(lowerName, raceRequest) {
				continue
			}
		}
		raceOptions = append(raceOptions, race)
	}

	return raceOptions
}
