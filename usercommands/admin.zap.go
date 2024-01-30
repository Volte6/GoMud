package usercommands

import (
	"fmt"

	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"

	"github.com/volte6/mud/users"
)

func Zap(rest string, userId int, cmdQueue util.CommandQueue) (util.MessageQueue, error) {

	response := NewUserCommandResponse(userId)

	// Load user details
	user := users.GetByUserId(userId)
	if user == nil { // Something went wrong. User not found.
		return response, fmt.Errorf("user %d not found", userId)
	}

	if user.Character.Aggro == nil || user.Character.Aggro.MobInstanceId == 0 {
		response.SendUserMessage(userId, "You are not in combat.", true)
		response.Handled = true
		return response, nil
	}

	mob := mobs.GetInstance(user.Character.Aggro.MobInstanceId)
	if mob == nil {
		response.SendUserMessage(userId, "Zap Mob not found.", true)
		response.Handled = true
		return response, nil
	}

	response.SendUserMessage(userId, fmt.Sprintf(`You zap <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, mob.Character.Name), true)
	response.SendRoomMessage(user.Character.RoomId, fmt.Sprintf(`<ansi fg="username">%s</ansi> zaps <ansi fg="mobname">%s</ansi> with a bolt of lightning!`, user.Character.Name, mob.Character.Name), true)
	mob.Character.Health = 1

	response.Handled = true
	return response, nil
}
