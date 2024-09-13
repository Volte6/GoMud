package usercommands

import (
	"fmt"

	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

type SkillsOptions struct {
	SkillList      map[string]int
	TrainingPoints int
	SkillCooldowns map[string]int
}

func Skills(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf(`user %d not found`, userId)
	}

	allSkills := user.Character.GetSkills()
	allCooldowns := map[string]int{}
	for skillName := range allSkills {
		allCooldowns[skillName] = user.Character.GetCooldown(skillName)
	}

	skillData := SkillsOptions{
		SkillList:      allSkills, // name to level
		SkillCooldowns: allCooldowns,
		TrainingPoints: user.Character.TrainingPoints,
	}

	skillTxt, _ := templates.Process("character/skills", skillData)
	response.SendUserMessage(userId, skillTxt, false)

	if rest == `extra` {
		response.SendUserMessage(userId, `<ansi fg="yellow">Cooldown Tracking:</ansi>`, true)
		for name, rnds := range user.Character.GetAllCooldowns() {
			response.SendUserMessage(userId, fmt.Sprintf(` <ansi fg="yellow">%s</ansi>: <ansi fg="red">%d</ansi>`, name, rnds), true)
		}
	}

	response.Handled = true
	return response, nil
}
