package scripting

import "github.com/volte6/mud/events"

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////

func SendUserMessage(userId int, message string) {
	if disableMessageQueue || userId == 0 {
		return
	}
	messageQueue.SendUserMessage(userId, message, true)
}

func SendRoomMessage(roomId int, message string, excludeIds ...int) {
	if disableMessageQueue {
		return
	}

	messageQueue.SendRoomMessage(roomId, message, true, excludeIds...)
}

func SendBroadcast(message string) {

	events.AddToQueue(events.Broadcast{Text: message + "\n"})

}
