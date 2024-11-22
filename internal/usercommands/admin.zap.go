package usercommands

import (
	"fmt"

	"github.com/volte6/gomud/internal/mobs"
	"github.com/volte6/gomud/internal/rooms"

	"github.com/volte6/gomud/internal/users"
)

func Zap(rest string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if rest != `` {

		playerId, mobId := room.FindByName(rest)

		if mobId > 0 {

			mob := mobs.GetInstance(mobId)
			if mob == nil {
				user.SendText("Zap Mob not found.")
				return true, nil
			}

			user.SendText(fmt.Sprintf(`You zap <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, mob.Character.Name))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> zaps <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, user.Character.Name, mob.Character.Name), user.UserId)
			mob.Character.Health = 1
			return true, nil
		}

		if playerId > 0 {
			if u := users.GetByUserId(playerId); u != nil {
				user.SendText(fmt.Sprintf(`You zap <ansi fg="username">%s</ansi> with a bolt of lightning!`, u.Character.Name))
				room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> zaps <ansi fg="username">%s</ansi> with a bolt of lightning!`, user.Character.Name, u.Character.Name), user.UserId, u.UserId)
				u.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> zaps you with a bolt of lightning!`, user.Character.Name))
				u.Character.Health = 1
				return true, nil
			}
		}

	}

	if user.Character.Aggro == nil || user.Character.Aggro.MobInstanceId == 0 {
		user.SendText("You are not in combat.")
		return true, nil
	}

	mob := mobs.GetInstance(user.Character.Aggro.MobInstanceId)
	if mob == nil {
		user.SendText("Zap Mob not found.")
		return true, nil
	}

	user.SendText(fmt.Sprintf(`You zap <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, mob.Character.Name))
	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> zaps <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, user.Character.Name, mob.Character.Name), user.UserId)
	mob.Character.Health = 1

	return true, nil
}
