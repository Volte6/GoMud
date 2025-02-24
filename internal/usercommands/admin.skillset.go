package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Skillset(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	// args should look like one of the following:
	// target buffId - put buff on target if in the room
	// buffId - put buff on self
	// search searchTerm - search for buff by name, display results
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.skillset", nil)
		user.SendText(infoOutput)

		user.SendText(`Skill Names:`)
		for _, name := range skills.GetAllSkillNames() {
			user.SendText(`  <ansi fg="skill">` + string(name) + `</ansi>`)
		}

		return true, nil
	}

	if args[0] == `all` {
		skillValueInt, _ := strconv.Atoi(args[1])

		for _, skillName := range skills.GetAllSkillNames() {
			user.Character.SetSkill(string(skillName), skillValueInt)
		}

		return true, nil
	}

	skillName := strings.ToLower(args[0])
	skillValueInt, _ := strconv.Atoi(args[1])

	found := skills.SkillExists(skillName)

	if found {
		user.Character.SetSkill(skillName, skillValueInt)
	} else {
		user.SendText(fmt.Sprintf(`Skill "%s" not found.`, skillName))
	}

	return true, nil
}
