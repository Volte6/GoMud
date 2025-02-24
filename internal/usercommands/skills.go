package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

type SkillsOptions struct {
	SkillList      map[string]int
	TrainingPoints int
	SkillCooldowns map[string]int
}

func Skills(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

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
	user.SendText(skillTxt)

	if rest == `extra` {
		user.SendText(`<ansi fg="yellow">Cooldown Tracking:</ansi>`)
		for name, rnds := range user.Character.GetAllCooldowns() {
			user.SendText(fmt.Sprintf(` <ansi fg="yellow">%s</ansi>: <ansi fg="red">%d</ansi>`, name, rnds))
		}
	}

	return true, nil
}
