package usercommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/util"

	"github.com/volte6/mud/users"
)

func Zap(rest string, userId int) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	// Load current room details
	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return response, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	if user.Character.Aggro == nil || user.Character.Aggro.MobInstanceId == 0 {
		user.SendText("You are not in combat.")
		response.Handled = true
		return response, nil
	}

	mob := mobs.GetInstance(user.Character.Aggro.MobInstanceId)
	if mob == nil {
		user.SendText("Zap Mob not found.")
		response.Handled = true
		return response, nil
	}

	user.SendText(fmt.Sprintf(`You zap <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, mob.Character.Name))
	room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> zaps <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, user.Character.Name, mob.Character.Name), userId)
	mob.Character.Health = 1

	response.Handled = true
	return response, nil
}
