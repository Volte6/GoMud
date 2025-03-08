package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/util"

	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

func Skillset(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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

	var targetUser *users.UserRecord = user

	foundUser, _ := room.FindByName(args[0])
	if foundUser > 0 {
		targetUser = users.GetByUserId(foundUser)
		args = args[1:]
	}

	if args[0] == `all` {
		skillValueInt, _ := strconv.Atoi(args[1])

		for _, skillName := range skills.GetAllSkillNames() {
			targetUser.Character.SetSkill(string(skillName), skillValueInt)
			targetUser.SendText(fmt.Sprintf(`Your "<ansi fg="skill">%s</ansi>" skill level has been set to <ansi fg="red">%d</ansi>.`, skillName, skillValueInt))
		}

		if targetUser.UserId != user.UserId {
			user.SendText("done.")
		}

		return true, nil
	}

	skillName := strings.ToLower(args[0])
	skillValueInt, _ := strconv.Atoi(args[1])

	found := skills.SkillExists(skillName)

	if found {
		targetUser.Character.SetSkill(skillName, skillValueInt)
		targetUser.SendText(fmt.Sprintf(`Your "<ansi fg="skill">%s</ansi>" skill level has been set to <ansi fg="red">%d</ansi>.`, skillName, skillValueInt))

		if targetUser.UserId != user.UserId {
			user.SendText("done.")
		}
	} else {
		targetUser.SendText(fmt.Sprintf(`Skill "%s" not found.`, skillName))
	}

	return true, nil
}
