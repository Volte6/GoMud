package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/mud/util"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
)

func Skillset(rest string, user *users.UserRecord) (bool, error) {

	// args should look like one of the following:
	// target buffId - put buff on target if in the room
	// buffId - put buff on self
	// search searchTerm - search for buff by name, display results
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.skillset", nil)
		user.SendText(infoOutput)
		return true, nil
	}

	skillName := strings.ToLower(args[0])
	skillValueInt, _ := strconv.Atoi(args[1])

	found := true

	if found {
		user.Character.SetSkill(skillName, skillValueInt)
	} else {
		user.SendText(fmt.Sprintf(`Skill "%s" not found.`, skillName))
	}

	return true, nil
}
