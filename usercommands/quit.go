package usercommands

import (
	"fmt"

	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
	"github.com/volte6/mud/util"
)

func Quit(rest string, userId int) (util.MessageQueue, error) {

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

	if user.Character.Aggro != nil {
		response.SendUserMessage(userId, "You're too busy to quit right now!", true)
		response.Handled = true
		return response, nil
	}

	events.AddToQueue(events.Buff{
		UserId:        userId,
		MobInstanceId: 0,
		BuffId:        0,
	})

	response.Handled = true
	return response, nil
}
