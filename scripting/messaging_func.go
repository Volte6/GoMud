package scripting

import (
	"github.com/volte6/mud/events"
	"github.com/volte6/mud/rooms"
	"github.com/volte6/mud/users"
)

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////

func SendUserMessage(userId int, message string) {
	if disableMessageQueue || userId == 0 {
		return
	}

	u := users.GetByUserId(userId)
	if u == nil {
		return
	}

	u.SendText(message)
}

func SendRoomMessage(roomId int, message string, excludeIds ...int) {
	if disableMessageQueue {
		return
	}

	r := rooms.LoadRoom(roomId)
	if r == nil {
		return
	}

	r.SendText(message, excludeIds...)
}

func SendRoomExitsMessage(roomId int, message string, excludeIds ...int) {
	if disableMessageQueue {
		return
	}

	r := rooms.LoadRoom(roomId)
	if r == nil {
		return
	}

	r.SendTextToExits(message, excludeIds...)

}

func SendBroadcast(message string) {

	events.AddToQueue(events.Broadcast{Text: message + "\n"})

}
