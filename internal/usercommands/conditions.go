package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/users"
)

func Conditions(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {

	type buffInfo struct {
		Name        string
		Description string
		RoundsLeft  int
		PermaBuff   bool
	}

	afflictions := []buffInfo{}

	charBuffs := user.Character.GetBuffs()
	for _, buff := range charBuffs {

		spec := buffs.GetBuffSpec(buff.BuffId)

		_, roundsLeft := buffs.GetDurations(buff, spec)

		newAffliction := buffInfo{
			Name:        spec.Name,
			Description: spec.Description,
			RoundsLeft:  roundsLeft,
			PermaBuff:   buff.PermaBuff,
		}
		newAffliction.Name, newAffliction.Description = spec.VisibleNameDesc()

		afflictions = append(afflictions, newAffliction)
	}

	tplTxt, _ := templates.Process("character/conditions", afflictions, user.UserId)
	user.SendText(tplTxt)

	return true, nil
}
