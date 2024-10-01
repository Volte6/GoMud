package usercommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"

	"github.com/volte6/mud/users"
)

func Zap(rest string, user *users.UserRecord) (bool, error) {

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
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
