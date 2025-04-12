package usercommands

import (
	"github.com/volte6/gomud/internal/buffs"
	"github.com/volte6/gomud/internal/events"
	"github.com/volte6/gomud/internal/rooms"
	"github.com/volte6/gomud/internal/templates"
	"github.com/volte6/gomud/internal/users"
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
