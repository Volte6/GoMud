package scripting

import (
	"github.com/dop251/goja"
	"github.com/volte6/mud/util"
)

var (
	disableMessageQueue = false
	messageQueue        util.MessageQueue
	commandQueue        util.CommandQueue
)

func setMessagingFunctions(vm *goja.Runtime) {

	vm.Set(`console`, newConsole(vm))
	vm.Set(`SendUserMessage`, SendUserMessage)
	vm.Set(`SendRoomMessage`, SendRoomMessage)

}

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
