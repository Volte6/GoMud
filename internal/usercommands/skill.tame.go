package usercommands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/characters"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/scripting"
	"github.com/volte6/gomud/internal/skills"
	"github.com/volte6/gomud/internal/spells"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
)

/*
Skill Tame
Level 1 - Tame up to 2 creatures
Level 2 - Tame up to 3 creatures
Level 3 - Tame up to 4 creatures
Level 4 - Tame up to 5 creatures
*/
func Tame(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	skillLevel := user.Character.GetSkillLevel(skills.Tame)
	if skillLevel == 0 {
		user.SendText("You don't know how to tame.")
		return true, errors.New(`you don't know how to tame`)
	}

	if rest == `list` || rest == `` {

		/*
			user.Character.MobMastery.SetTame(1, 87)
			user.Character.MobMastery.SetTame(54, 50)
			user.Character.MobMastery.SetTame(55, 40)
		*/
		headers := []string{`Name`, `Proficiency`}
		rows := [][]string{}

		for mobId, proficiency := range user.Character.MobMastery.GetAllTame() {

			mobInfo := mobs.GetMobSpec(mobs.MobId(mobId))
			if mobInfo == nil {
				continue
			}

			rows = append(rows, []string{
				//mobInfo.Character.Name,
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi>`, mobInfo.Character.Name),
				strconv.Itoa(proficiency) + `%`,
			})
		}

		tameTableData := templates.GetTable(`Your taming proficiency`, headers, rows)
		tplTxt, _ := templates.Process("tables/generic", tameTableData, user.UserId)
		user.SendText(tplTxt)
		user.SendText(`<ansi fg="command">help tame</ansi> to find out more.`)

		return true, nil
	}

	// valid peep targets are: mobs, players
	_, mobId := room.FindByName(rest)

	if mobId > 0 {

		if mob := mobs.GetInstance(mobId); mob != nil {

			if mob.Character.IsCharmed(user.UserId) {
				user.SendText("They are already charmed.")
				return true, errors.New(`they are already charmed`)
			}

			// Set spell Aid
			spellAggro := characters.SpellAggroInfo{
				SpellId:              `tameskill`,
				SpellRest:            ``,
				TargetUserIds:        []int{},
				TargetMobInstanceIds: []int{mobId},
			}

			continueCasting := true
			if handled, err := scripting.TrySpellScriptEvent(`onCast`, user.UserId, 0, spellAggro); err == nil {
				continueCasting = handled
			}

			if continueCasting {
				spellInfo := spells.GetSpell(`tameskill`)
				user.Character.CancelBuffsWithFlag(buffs.Hidden)
				user.Character.SetCast(spellInfo.WaitRounds, spellAggro)
			}

			return true, nil

		}

	}

	user.SendText("You don't see that here.")

	return true, errors.New(`you don't see that here`)

}
