package usercommands

import (
	"math"

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
	}

	afflictions := []buffInfo{}

	charBuffs := user.Character.GetBuffs()
	for _, buff := range charBuffs {

		spec := buffs.GetBuffSpec(buff.BuffId)
		totalRounds := int(math.Ceil(float64(buff.TriggersLeft) * float64(spec.RoundInterval)))

		newAffliction := buffInfo{
			Name:        spec.Name,
			Description: spec.Description,
			RoundsLeft:  totalRounds - (buff.RoundCounter),
		}

		if spec.Secret {
			newAffliction.Name = "Mysterious Affliction"
			newAffliction.Description = "Unknown"
		}

		afflictions = append(afflictions, newAffliction)
	}

	tplTxt, _ := templates.Process("character/conditions", afflictions)
	user.SendText(tplTxt)

	return true, nil
}
