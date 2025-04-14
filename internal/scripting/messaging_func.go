package scripting

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
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

	message = userTextWrap.Wrap(message)

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

	message = roomTextWrap.Wrap(message)

	r.SendText(message, excludeIds...)
}

func SendRoomExitsMessage(roomId int, message string, isQuiet bool, excludeUserIds ...int) {
	if disableMessageQueue {
		return
	}

	r := rooms.LoadRoom(roomId)
	if r == nil {
		return
	}

	message = roomTextWrap.Wrap(message)

	r.SendTextToExits(message, isQuiet, excludeUserIds...)

}

func SendBroadcast(message string) {

	message = roomTextWrap.Wrap(message)

	events.AddToQueue(events.Broadcast{Text: message + "\n"})

}
