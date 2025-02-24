package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/colorpatterns"
	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"

	"github.com/volte6/gomud/internal/users"
)

func Paz(rest string, user *users.UserRecord, room *rooms.Room, flags UserCommandFlag) (bool, error) {

	beamOfLight := colorpatterns.ApplyColorPattern(`beam of light`, `rainbow`)

	if rest != `` {

		playerId, mobId := room.FindByName(rest)

		if mobId > 0 {

			mob := mobs.GetInstance(mobId)
			if mob == nil {
				user.SendText("Paz Mob not found.")
				return true, nil
			}

			user.SendText(fmt.Sprintf(`You illuminate <ansi fg="mobname">%s</ansi> with a %s!`, mob.Character.Name, beamOfLight))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> illuminates <ansi fg="mobname">%s</ansi> with a %s!`, user.Character.Name, mob.Character.Name, beamOfLight), user.UserId)

			mob.Character.Health = mob.Character.HealthMax.Value
			mob.Character.Mana = mob.Character.ManaMax.Value

			return true, nil
		}

		if playerId > 0 {
			if u := users.GetByUserId(playerId); u != nil {
				user.SendText(fmt.Sprintf(`You illuminate <ansi fg="username">%s</ansi> with a %s!`, u.Character.Name, beamOfLight))
				room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> illuminates <ansi fg="username">%s</ansi> with a %s!`, user.Character.Name, u.Character.Name, beamOfLight), user.UserId, u.UserId)
				u.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> illuminates you with a %s!`, user.Character.Name, beamOfLight))

				u.Character.Health = u.Character.HealthMax.Value
				u.Character.Mana = u.Character.ManaMax.Value

				return true, nil
			}
		}

	}

	user.SendText(`You paz yourself with a ` + beamOfLight + `!`)
	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> illuminates <ansi fg="username">%s</ansi> with a %s!`, user.Character.Name, user.Character.Name, beamOfLight), user.UserId)

	user.Character.Health = user.Character.HealthMax.Value
	user.Character.Mana = user.Character.ManaMax.Value

	return true, nil
}
