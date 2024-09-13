package usercommands

import (
	"errors"
	"fmt"

	"github.com/volte6/mud/buffs"
	"github.com/volte6/mud/characters"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/scripting"
	"github.com/volte6/mud/skills"
	"github.com/volte6/mud/spells"
	"github.com/volte6/mud/templates"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

/*
Skill Tame
Level 1 - Tame up to 2 creatures
Level 2 - Tame up to 3 creatures
Level 3 - Tame up to 4 creatures
Level 4 - Tame up to 5 creatures
*/
func Tame(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	skillLevel := user.Character.GetSkillLevel(skills.Tame)
	creatureTameSkill := make(map[string]int)

	for _, creatureName := range user.Character.GetMiscDataKeys(`tameskill-`) {

		skillValue := user.Character.GetMiscData(`tameskill-` + creatureName)

		if skillValueInt, ok := skillValue.(int); ok {
			creatureTameSkill[creatureName] = skillValueInt
		} else {
			creatureTameSkill[creatureName] = 0
		}

	}

	if skillLevel == 0 {
		response.SendUserMessage(userId, "You don't know how to tame.", true)
		response.Handled = true
		return response, errors.New(`you don't know how to tame`)
	}

	if len(rest) == 0 {

		headers := []string{"Creature", "Proficiency"}

		rows := [][]string{}

		for creatureName, modProficiency := range creatureTameSkill {
			rows = append(rows, []string{creatureName, fmt.Sprintf("%d", modProficiency)})
		}

		onlineTableData := templates.GetTable(`Your taming proficiency`, headers, rows)
		tplTxt, _ := templates.Process("tables/generic", onlineTableData)
		response.SendUserMessage(userId, tplTxt, true)

		response.SendUserMessage(userId, `<ansi fg="command">help tame</ansi> to find out more.`, true)

		response.Handled = true
		return response, nil
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	// valid peep targets are: mobs, players
	_, mobId := room.FindByName(rest)

	if mobId > 0 {

		if mob := mobs.GetInstance(mobId); mob != nil {

			if mob.Character.IsCharmed(userId) {
				response.SendUserMessage(userId, "They are already charmed.", true)
				response.Handled = true
				return response, errors.New(`they are already charmed`)
			}

			// Set spell Aid
			spellAggro := characters.SpellAggroInfo{
				SpellId:              `tameskill`,
				SpellRest:            ``,
				TargetUserIds:        []int{},
				TargetMobInstanceIds: []int{mobId},
			}

			continueCasting := true
			if res, err := scripting.TrySpellScriptEvent(`onCast`, userId, 0, spellAggro); err == nil {
				response.AbsorbMessages(res)
				continueCasting = res.Handled
			}

			if continueCasting {
				spellInfo := spells.GetSpell(`tameskill`)
				user.Character.CancelBuffsWithFlag(buffs.Hidden)
				user.Character.SetCast(spellInfo.WaitRounds, spellAggro)
			}

			response.Handled = true
			return response, nil

		}

	}

	response.SendUserMessage(userId, "You don't see that here.", true)
	response.Handled = true
	return response, errors.New(`you don't see that here`)

}
