package scripting

// ////////////////////////////////////////////////////////
//
// # These functions get exported to the scripting engine
//
// ////////////////////////////////////////////////////////

func SendUserMessage(userId int, message string) {
	if disableMessageQueue {
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
	commandQueue.Broadcast(message + "\n")
}
